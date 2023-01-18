package client

import (
	"k8s.io/client-go/discovery"
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

	// NewDynamicClient
	clientSet.NewDynamicClient = func() *DynamicClientWrap {
		return NewDynamicClient(restConfig)
	}

	// 예:  schema.GroupVersionResource{Group: "networking.istio.io", Version: "v1alpha3", Resource: "virtualservices"}
	clientSet.NewDynamicClientSchema = func(group string, version string, resource string) *DynamicClientWrap {
		return NewDynamicClientSchema(restConfig, group, version, resource)
	}

	return clientSet
}
