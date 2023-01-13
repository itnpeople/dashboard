package config

import (
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/kore3lab/dashboard/backend/pkg/auth"
	"github.com/kore3lab/dashboard/backend/pkg/client"
	"github.com/kore3lab/dashboard/backend/pkg/kubeconfig"
)

const (
	IN_CLUSTER_NAME             = "kubernetes@in-cluster"
	KubeConfigStrategyFile      = "file"
	KubeConfigStrategyConfigmap = "configmap"
)

type conf struct {
	MetricsScraperUrl string                     // metrics scraper
	TerminalUrl       string                     // terminal service Url
	AuthConfig        *auth.AuthenticatorOptions // auth-config
	KubeContexts      *kubeconfig.KubeContexts   // kubeconfig context
	//KubeConfig        *kubeConfig      // kubeconfig file
}

type kubeConfig struct {
	Strategy string            `json:"strategy"` // (file,configmap)
	Data     map[string]string // file, configmap, namespace, filename
}

type ClientSet struct {
	Name                       string
	RESTConfig                 *rest.Config
	NewCumulativeMetricsClient func() *client.CumulativeMetricsClient
	NewMetricsClient           func() (*versioned.Clientset, error)
	NewKubernetesClient        func() (*kubernetes.Clientset, error)
	NewDiscoveryClient         func() (*discovery.DiscoveryClient, error)
	NewDynamicClient           func() (*client.DynamicClient, error)
	NewDynamicClientSchema     func(group string, version string, resource string) (*client.DynamicClient, error)
}

type kubeCluster struct {
	KubeConfig         *api.Config                         // kubeconfig file
	IsRunningInCluster bool                                // cluster 지정이 안되어 있어서 in-cluster 를 default cluster 로 자동 지정된 경우
	InCluster          *ClientSet                          // in-cluter
	DefaultContext     string                              // kubeconfig file - default context
	Save               func() error                        // kubeconfig save
	ClusterNames       []string                            // context list
	clients            map[string]*ClientSet               // clusters client (rest.Config)
	Client             func(ct string) (*ClientSet, error) // get cluster client
}
