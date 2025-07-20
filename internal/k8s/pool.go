package k8s

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/strowk/foxy-contexts/pkg/session"
	"github.com/strowk/mcp-k8s-go/internal/k8s/list_mapping"
	"go.uber.org/fx"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/cache"
)

var (
	ErrInformerSyncTimeout = errors.New("informer took too long to sync")
	ErrUnknownSyncError    = errors.New("unknown sync error")
)

// ClientPool is a pool of Kubernetes clientsets and informers
// that can be used to interact with Kubernetes resources.
//
// It is thread-safe and can be used from multiple goroutines.
// It caches the clientsets and informers for each context
// to avoid creating them multiple times.
type ClientPool interface {
	GetClientset(ctx context.Context, k8sContext string) (kubernetes.Interface, error)
	GetDynamicClient(k8sContext string) (dynamic.Interface, error)
	GetInformer(
		ctx context.Context,
		namespace string,
		k8sCtx string,
		kind string,
		group string,
		version string,
	) (informers.GenericInformer, error)
	GetListMapping(k8sCtx, kind, group, version string) list_mapping.ListMapping
}

type resolvedResource struct {
	gvk     *schema.GroupVersionKind
	mapping *meta.RESTMapping

	namespaceToInformer map[string]informers.GenericInformer

	// informer    informers.GenericInformer
	listMapping list_mapping.ListMapping
}

type pool struct {
	clients        map[string]kubernetes.Interface
	dynamicClients map[string]dynamic.Interface

	getClientsetMutex     *sync.Mutex
	getDynamicClientMutex *sync.Mutex

	keyToResource map[string]*resolvedResource
	gvkToResource map[schema.GroupVersionKind]*resolvedResource

	getInformerMutex *sync.Mutex

	listMappingResolvers []list_mapping.ListMappingResolver

	fcSessionManager *session.SessionManager

	stopChan chan struct{}
}

func NewClientPool(
	listMappingResolvers []list_mapping.ListMappingResolver,
	sessionManager *session.SessionManager,
	lifecycle fx.Lifecycle,
) ClientPool {
	p := &pool{
		clients:           make(map[string]kubernetes.Interface),
		getClientsetMutex: &sync.Mutex{},

		dynamicClients:        make(map[string]dynamic.Interface),
		getDynamicClientMutex: &sync.Mutex{},

		keyToResource:    make(map[string]*resolvedResource),
		gvkToResource:    make(map[schema.GroupVersionKind]*resolvedResource),
		getInformerMutex: &sync.Mutex{},

		listMappingResolvers: listMappingResolvers,

		fcSessionManager: sessionManager,

		stopChan: make(chan struct{}),
	}

	lifecycle.Append(
		fx.StopHook(func(ctx context.Context) error {
			close(p.stopChan)
			return nil
		}),
	)

	return p
}

func getResourceKey(
	k8sCtx string,
	kind string,
	group string,
	version string,
) string {
	return fmt.Sprintf("%s/%s/%s/%s", k8sCtx, kind, group, version)
}

func (p *pool) GetInformer(
	ctx context.Context,
	namespace string,
	k8sCtx string,
	kind string,
	group string,
	version string,
) (informers.GenericInformer, error) {
	// creating informer needs to be thread-safe to avoid creating
	// multiple informers for the same resource
	p.getInformerMutex.Lock()
	defer p.getInformerMutex.Unlock()

	// this looks up if we have a resource with informer already
	// for exactly the same requested context and "lookup" gvk
	key := getResourceKey(k8sCtx, kind, group, version)
	if res, ok := p.keyToResource[key]; ok {
		if inf, ok := res.namespaceToInformer[namespace]; ok {
			return inf, nil
		}
	}

	// if not, then we resolve gvk and mapping from what server has
	res, err := p.resolve(ctx, k8sCtx, kind, group, version)
	if err != nil {
		return nil, err
	}

	// it is still possible for this resource to be known already
	// just with different key, so we check if we have it already,
	// now by canonical resolved gvk
	alreadySetupResource, ok := p.gvkToResource[*res.gvk]
	if ok {
		// rememeber that this key is resolved to already known resource
		p.keyToResource[key] = alreadySetupResource
		// and we can return the informer for it
		if inf, ok := alreadySetupResource.namespaceToInformer[namespace]; ok {
			return inf, nil
		}
	}

	// if not, then we setup informer and cache it
	inf, err := res.setupInformer(k8sCtx, namespace, p.fcSessionManager, p.stopChan, ctx)
	if err != nil {
		return nil, err
	}
	p.keyToResource[key] = res
	p.gvkToResource[*res.gvk] = res
	return inf, nil
}

func (p *pool) resolve(
	ctx context.Context,
	k8sCtx string,
	kind string,
	group string,
	version string,
) (*resolvedResource, error) {
	clientset, err := p.GetClientset(ctx, k8sCtx)
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(clientset.Discovery()))

	serverPreferredResources, err := clientset.Discovery().ServerPreferredResources()
	if serverPreferredResources == nil && err != nil {
		return nil, err
	}

	lookupGvk := schema.GroupVersionKind{
		Group:   strings.ToLower(group),
		Version: strings.ToLower(version),
		Kind:    strings.ToLower(kind),
	}

	var resolvedGvk *schema.GroupVersionKind
	var resolvedMapping *meta.RESTMapping

lookingForResource:
	for _, r := range serverPreferredResources {
		for _, apiResource := range r.APIResources {
			resourceKind := apiResource.Kind
			resourceGroup := apiResource.Group
			resourceVersion := apiResource.Version
			if resourceGroup == "" || resourceVersion == "" {
				// some resources have empty group or version, which is then present in the containing resource list
				// for example: apps/v1
				// we need to set the group and version to the one from the containing resource list
				split := strings.SplitN(r.GroupVersion, "/", 2)
				if len(split) == 2 {
					resourceGroup = split[0]
					resourceVersion = split[1]
				} else {
					resourceVersion = r.GroupVersion
				}
			}

			if strings.EqualFold(apiResource.Kind, lookupGvk.Kind) {
				// some resources cannot have correct RESTMapping, but we need to create it
				// to check if the requested resource matches what we have found in the server
				// , but we can at first at least look if the kind matches what we are looking for
				gvk := schema.GroupVersionKind{
					Group:   resourceGroup,
					Version: resourceVersion,
					Kind:    resourceKind,
				}

				mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
				if err != nil {
					return nil, fmt.Errorf("failed to get rest mapping for %s: %w", gvk.String(), err)
				}

				if strings.EqualFold(mapping.GroupVersionKind.Kind, lookupGvk.Kind) &&
					// if group or version were not specified, we would ignore them when matching
					// and this would simply match the first resource with matching kind
					(lookupGvk.Group == "" || strings.EqualFold(mapping.GroupVersionKind.Group, lookupGvk.Group)) &&
					(lookupGvk.Version == "" || strings.EqualFold(mapping.GroupVersionKind.Version, lookupGvk.Version)) {

					resolvedGvk = &gvk
					resolvedMapping = mapping
					break lookingForResource
				}
			}
		}
	}

	if resolvedGvk == nil {
		return nil, fmt.Errorf("resource %s/%s/%s not found", group, version, kind)
	}

	log.Printf("Resolved resource %s/%s/%s to GVK %s", group, version, kind, resolvedGvk.String())

	return &resolvedResource{
		gvk:                 resolvedGvk,
		mapping:             resolvedMapping,
		namespaceToInformer: make(map[string]informers.GenericInformer),
	}, nil
}

func (res *resolvedResource) setupInformer(
	k8sCtx string,
	namespace string,
	sessionManager *session.SessionManager,
	stopChan chan struct{},
	ctx context.Context,
) (informers.GenericInformer, error) {
	kubeConfig := GetKubeConfigForContext(k8sCtx)
	restConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	session, _ := sessionManager.GetSessionFromContext(ctx)
	if session != nil && session.AuthUserId != "" {
		restConfig.BearerToken = session.AuthUserId

		// Clean out thing that are not needed, but could be set in kubeconfig
		restConfig.KeyData = nil
		restConfig.CertData = nil
		restConfig.Username = ""
		restConfig.Password = ""
		restConfig.Impersonate = rest.ImpersonationConfig{}

		log.Printf("Using OIDC token to setup informer for context %s", k8sCtx)
	}

	dynClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		dynClient,
		10*time.Minute,
		namespace,
		nil,
	)
	informer := factory.ForResource(res.mapping.Resource)

	var syncErr error
	syncErrWriteMux := &sync.Mutex{}

	localStopChan := make(chan struct{})
	go func() {
		select {
		case <-stopChan:
			close(localStopChan)
		case <-localStopChan:
			// already closed, nothing to do
			return
		case <-time.After(5 * time.Second):
			log.Printf(
				"Informer for resource %s/%s/%s is taking too long to start, stopping it",
				res.mapping.GroupVersionKind.Group,
				res.mapping.GroupVersionKind.Version,
				res.mapping.GroupVersionKind.Kind,
			)
			syncErrWriteMux.Lock()
			defer syncErrWriteMux.Unlock()
			if syncErr == nil {
				syncErr = ErrInformerSyncTimeout
			}
			close(localStopChan)
		}
	}()

	err = informer.Informer().SetWatchErrorHandler(func(r *cache.Reflector, err error) {
		syncErrWriteMux.Lock()
		defer syncErrWriteMux.Unlock()
		syncErr = err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to set watch error handler for informer: %w", err)
	}

	go informer.Informer().Run(localStopChan)
	isSynced := cache.WaitForCacheSync(localStopChan, informer.Informer().HasSynced)
	if !isSynced {
		if syncErr == nil {
			syncErr = ErrUnknownSyncError
		}
		return nil, fmt.Errorf(
			"informer for resource %s/%s/%s could not sync: %w",
			res.mapping.GroupVersionKind.Group,
			res.mapping.GroupVersionKind.Version,
			res.mapping.GroupVersionKind.Kind,
			syncErr,
		)
	} else {
		err = informer.Informer().SetWatchErrorHandler(func(r *cache.Reflector, err error) {
			log.Printf("Informer for resource %s/%s/%s encountered an error: %v",
				res.mapping.GroupVersionKind.Group,
				res.mapping.GroupVersionKind.Version,
				res.mapping.GroupVersionKind.Kind,
				err,
			)
		})
		if err != nil {
			return nil, fmt.Errorf("failed to set watch error handler for informer: %w", err)
		}
	}
	res.namespaceToInformer[namespace] = informer
	return informer, nil
}

func (p *pool) GetClientset(
	ctx context.Context,
	k8sContext string,
) (kubernetes.Interface, error) {
	p.getClientsetMutex.Lock()
	defer p.getClientsetMutex.Unlock()

	var effectiveContext string
	if k8sContext == "" {
		var err error
		effectiveContext, err = GetCurrentContext()
		if err != nil {
			return nil, err
		}
	} else {
		effectiveContext = k8sContext
	}

	if !IsContextAllowed(effectiveContext) {
		return nil, fmt.Errorf("context %s is not allowed", effectiveContext)
	}

	key := effectiveContext
	if client, ok := p.clients[key]; ok {
		return client, nil
	}

	session, _ := p.fcSessionManager.GetSessionFromContext(ctx)
	var client kubernetes.Interface
	var err error
	if session == nil || session.AuthUserId == "" {
		client, err = getClientset(k8sContext)
		if err != nil {
			return nil, fmt.Errorf("failed to get clientset for context %s: %w", k8sContext, err)
		}
	} else {
		log.Printf("Using OIDC token to setup client set for context %s", k8sContext)
		client, err = getOIDCAuthenticatedClientSet(k8sContext, session.AuthUserId)
		key = fmt.Sprintf("%s/%s", effectiveContext, session.SessionID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get OIDC authenticated clientset: %w", err)
		}
	}

	p.clients[key] = client
	return client, nil
}

func getOIDCAuthenticatedClientSet(k8sContext string, token string) (kubernetes.Interface, error) {
	// this should override kubeconfig with using the token
	kubeConfig := GetKubeConfigForContext(k8sContext)
	restConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig client config: %w", err)
	}

	restConfig.BearerToken = token

	// Clean out thing that are not needed, but could be set in kubeconfig
	restConfig.KeyData = nil
	restConfig.CertData = nil
	restConfig.Username = ""
	restConfig.Password = ""
	restConfig.Impersonate = rest.ImpersonationConfig{}

	// TODO: support refreshing..

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}
	return clientset, nil
}

func getClientset(
	k8sContext string,
) (kubernetes.Interface, error) {
	kubeConfig := GetKubeConfigForContext(k8sContext)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func (p *pool) GetDynamicClient(k8sContext string) (dynamic.Interface, error) {
	p.getDynamicClientMutex.Lock()
	defer p.getDynamicClientMutex.Unlock()

	if k8sContext == "" {
		k8sContext = "default"
	}

	if client, ok := p.dynamicClients[k8sContext]; ok {
		return client, nil
	}
	kubeConfig := GetKubeConfigForContext(k8sContext)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	p.dynamicClients[k8sContext] = client
	return client, nil
}
