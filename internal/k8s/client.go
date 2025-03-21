package k8s

import (
	"github.com/strowk/mcp-k8s-go/internal/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func GetKubeConfig() clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, nil)
	return kubeConfig
}

func GetKubeConfigForContext(k8sContext string) clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	if k8sContext == "" {
		configOverrides = nil
	} else {
		configOverrides.CurrentContext = k8sContext
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)
}

func GetCurrentContext() (string, error) {
	kubeConfig := GetKubeConfig()
	config, err := kubeConfig.RawConfig()
	if err != nil {
		return "", err
	}
	return config.CurrentContext, nil
}

func GetKubeClientset() (*kubernetes.Clientset, error) {
	kubeConfig := GetKubeConfig()

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

// IsContextAllowed checks if a context is allowed based on the configuration
func IsContextAllowed(contextName string) bool {
	return config.IsContextAllowed(contextName)
}
