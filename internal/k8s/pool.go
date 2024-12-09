package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type ClientPool interface {
	GetClientset(k8sContext string) (kubernetes.Interface, error)
}

type pool struct {
	clients map[string]kubernetes.Interface
}

func NewClientPool() ClientPool {
	return &pool{
		clients: make(map[string]kubernetes.Interface),
	}
}

func (p *pool) GetClientset(k8sContext string) (kubernetes.Interface, error) {
	key := k8sContext
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
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	if k8sContext == "" {
		configOverrides = nil
	} else {
		configOverrides.CurrentContext = k8sContext
	}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

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
