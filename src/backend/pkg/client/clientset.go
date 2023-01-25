package client

import (
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

func NewClientSet(restConfig *rest.Config) *ClientSet {

	clientSet := &ClientSet{}

	// NewDynamicClient
	clientSet.NewMetricsClient = func() (*versioned.Clientset, error) {
		return versioned.NewForConfig(restConfig)
	}
	// NewKubernetesClient
	clientSet.NewKubernetesClient = func() (*kubernetes.Clientset, error) {
		return kubernetes.NewForConfig(restConfig)
	}

	// NewDiscoveryClient
	clientSet.NewDiscoveryClient = func() (*discovery.DiscoveryClient, error) {
		return discovery.NewDiscoveryClientForConfig(restConfig)
	}

	// NewDynamicClint
	clientSet.NewDynamicClint = func() (*dynamic.DynamicClient, error) {
		return dynamic.NewForConfig(restConfig)
	}

	return clientSet
}
