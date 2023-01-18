package client

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

// type resourceVerber struct {
type DynamicClientWrap struct {
	config       *rest.Config
	resource     schema.GroupVersionResource
	namespace    string
	namespaceSet bool
}

type ClientSet struct {
	NewMetricsClient       func() (*versioned.Clientset, error)
	NewKubernetesClient    func() (*kubernetes.Clientset, error)
	NewDiscoveryClient     func() (*discovery.DiscoveryClient, error)
	NewDynamicClient       func() *DynamicClientWrap
	NewDynamicClientSchema func(group string, version string, resource string) *DynamicClientWrap
}

type CumulativeMetricsClient struct {
	Get func(selector CumulativeMetricsResourceSelector) ([]CumulativeMetricUnit, error)
}

type CumulativeMetricsResourceSelector struct {
	Node      string
	Namespace string
	Pods      []string
	Function  string
}

type CumulativeMetricUnit struct {
	CPU       int64  `json:"cpu"`
	Memory    int64  `json:"memory"`
	Timestamp string `json:"timestamp"`
}
