package client

import (
	"io"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

// type resourceVerber struct {
type KubeExecutor struct {
	namespace string
	POST      func(payload io.Reader, isUpdate bool) (output *unstructured.Unstructured, err error)
}

type ClientSet struct {
	NewMetricsClient    func() (*versioned.Clientset, error)
	NewKubernetesClient func() (*kubernetes.Clientset, error)
	NewDiscoveryClient  func() (*discovery.DiscoveryClient, error)
	NewDynamicClint     func() (*dynamic.DynamicClient, error)
	NewKubeExecutor     func() *KubeExecutor
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
