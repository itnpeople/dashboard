package kubeconfig

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

const (
	IN_CLUSTER_NAME             string = "kubernetes@in-cluster"
	KubeConfigStrategyFile      string = "file"
	KubeConfigStrategyConfigmap string = "configmap"
)

type KubeConfigOptions struct {
	Strategy  string `json:"strategy"` // (file,configmap)
	ConfigMap string `json:"configmap"`
	Namespace string `json:"namespace"`
	Filename  string `json:"filename"`
}

type KubeConfigProvider struct {
	*KubeConfigOptions
	APIConfig  *api.Config
	Read       func() error
	Write      func() error
	IsModified func() bool
}

type KubeContexts struct {
	RESTConfigs     map[string]*rest.Config
	CurrentContext  string
	Client          func(ctx string) (*kubernetes.Clientset, error)
	VersionedClient func(ctx string) (*versioned.Clientset, error)
	DynamicClient   func(ctx string) (*dynamic.DynamicClient, error)
}
