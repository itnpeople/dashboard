package clusters

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/kore3lab/dashboard/backend/pkg/client"
)

const (
	IN_CLUSTER_NAME             string             = "kubernetes@in-cluster"
	KubeConfigStrategyFile      KubeConfigStrategy = "file"
	KubeConfigStrategyConfigmap KubeConfigStrategy = "configmap"
)

type KubeConfigStrategy string

type KubeConfigOptions struct {
	Strategy  KubeConfigStrategy `json:"strategy"` // (file,configmap)
	ConfigMap string             `json:"configmap"`
	Namespace string             `json:"namespace"` // if strategy=configmap
	Filename  string             `json:"filename"`  // if strategy=file
}

type KubeConfigProvider struct {
	options    *KubeConfigOptions
	apiConfig  *api.Config
	read       func() error
	write      func() error
	isModified func() bool
}

type KubeClusters struct {
	*KubeConfigProvider
	restConfigs     map[string]*rest.Config
	CurrentCluster  string
	NewClientSet    func(clusterName string) (*client.ClientSet, error)
	AddCluster      func(params *KubeConfig) error
	RemoveCluster   func(clusterName string) error
	GetClusterNames func() []string
}

type KubeConfig struct {
	Name    string `json:"name"`
	Cluster struct {
		Server                   string `json:"server"`
		CertificateAuthorityData string `json:"certificate-authority-data"`
	} `json:"cluster"`
	User struct {
		ClientCertificateData string `json:"client-certificate-data"`
		ClientKeyData         string `json:"client-key-data"`
		Token                 string `json:"token"`
	} `json:"user"`
}
