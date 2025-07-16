package k8s

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/strowk/mcp-k8s-go/internal/k8s/list_mapping"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/cache"
)

// ClientPool is a pool of Kubernetes clientsets and informers
// that can be used to interact with Kubernetes resources.
//
// It is thread-safe and can be used from multiple goroutines.
// It caches the clientsets and informers for each context
// to avoid creating them multiple times.
type ClientPool interface {
	GetClientset(k8sContext string) (kubernetes.Interface, error)
	GetDynamicClient(k8sContext string) (dynamic.Interface, error)
	GetInformer(
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

	informer    informers.GenericInformer
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
}

func NewClientPool(listMappingResolvers []list_mapping.ListMappingResolver) ClientPool {
	return &pool{
		clients:           make(map[string]kubernetes.Interface),
		getClientsetMutex: &sync.Mutex{},

		dynamicClients:        make(map[string]dynamic.Interface),
		getDynamicClientMutex: &sync.Mutex{},

		keyToResource:    make(map[string]*resolvedResource),
		gvkToResource:    make(map[schema.GroupVersionKind]*resolvedResource),
		getInformerMutex: &sync.Mutex{},

		listMappingResolvers: listMappingResolvers,
	}
}

func (p *pool) GetInformer(
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
	key := fmt.Sprintf("%s/%s/%s/%s", k8sCtx, kind, group, version)
	if res, ok := p.keyToResource[key]; ok {
		return res.informer, nil
	}

	// if not, then we resolve gvk and mapping from what server has
	res, err := p.resolve(k8sCtx, kind, group, version)
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
		return alreadySetupResource.informer, nil
	}

	// if not, then we setup informer and cache it
	err = res.setupInformer(k8sCtx)
	if err != nil {
		return nil, err
	}
	p.keyToResource[key] = res
	p.gvkToResource[*res.gvk] = res
	return res.informer, nil
}

func (p *pool) resolve(
	k8sCtx string,
	kind string,
	group string,
	version string,
) (*resolvedResource, error) {
	clientset, err := p.GetClientset(k8sCtx)
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

	return &resolvedResource{
		gvk:     resolvedGvk,
		mapping: resolvedMapping,
	}, nil
}

func (res *resolvedResource) setupInformer(
	k8sCtx string,
) error {
	cfg := GetKubeConfigForContext(k8sCtx)
	restConfig, err := cfg.ClientConfig()
	if err != nil {
		return err
	}

	dynClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynClient, 10*time.Minute, metav1.NamespaceAll, nil)
	informer := factory.ForResource(res.mapping.Resource)
	go informer.Informer().Run(context.Background().Done())
	isSynced := cache.WaitForCacheSync(context.Background().Done(), informer.Informer().HasSynced)
	if !isSynced {
		return fmt.Errorf("informer for resource %s/%s/%s is not synced", res.mapping.GroupVersionKind.Group, res.mapping.GroupVersionKind.Version, res.mapping.GroupVersionKind.Kind)
	}
	res.informer = informer
	return nil
}

func (p *pool) GetClientset(k8sContext string) (kubernetes.Interface, error) {
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

	client, err := getClientset(k8sContext)
	if err != nil {
		return nil, err
	}

	p.clients[key] = client
	return client, nil
}

func getClientset(k8sContext string) (kubernetes.Interface, error) {
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
