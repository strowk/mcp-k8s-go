package k8s

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
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
