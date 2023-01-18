package clusters

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/kore3lab/dashboard/backend/pkg/client"
	"github.com/kore3lab/dashboard/backend/pkg/lang"
)

func NewKubeClusters(opts string) (*KubeClusters, error) {

	// populate a KubeConfigProvider
	var provider *KubeConfigProvider
	var options *KubeConfigOptions
	var err error

	if options, err = getKubeConfigOptions(opts); err != nil {
		return nil, err
	} else {
		if provider, err = NewKubeConfigProvider(options); err != nil {
			return nil, err
		} else {
			if err = provider.read(); err != nil {
				return nil, err
			}
		}
	}

	// populate a KubeClusters
	clusters := &KubeClusters{KubeConfigProvider: provider}
	clusters.restConfigs = make(map[string]*rest.Config)
	if apiConfig := provider.apiConfig; apiConfig == nil {
		if inClusterConfig, err := rest.InClusterConfig(); err == nil {
			clusters.restConfigs[IN_CLUSTER_NAME] = inClusterConfig
			clusters.CurrentCluster = IN_CLUSTER_NAME
		}
	} else {
		clusters.CurrentCluster = apiConfig.CurrentContext
		for ctx := range apiConfig.Contexts {
			if restConfig, err := clientcmd.NewNonInteractiveClientConfig(*apiConfig, ctx, &clientcmd.ConfigOverrides{}, nil).ClientConfig(); err == nil {
				clusters.restConfigs[ctx] = restConfig
				if clusters.CurrentCluster == "" {
					clusters.CurrentCluster = ctx
				}
			}
		}
	}

	// kubernetes.Clientset
	clusters.NewClientSet = func(ctx string) (*client.ClientSet, error) {

		if provider.isModified() {
			if err := provider.read(); err != nil {
				return nil, err
			}
		}
		if restConfig, ok := clusters.restConfigs[ctx]; ok {
			return client.NewClientSet(restConfig), nil
		} else {
			return nil, errors.New(fmt.Sprintf("The Kubernetes-Client could not be created because the context was not found. (context=%s)", ctx))
		}
	}

	clusters.AddCluster = func(params *KubeConfig) error {

		// create objects
		cluster := &api.Cluster{}
		context := &api.Context{}
		user := &api.AuthInfo{}

		// context, cluster, user 이름 중복 회피
		if provider.apiConfig != nil {
			if _, exist := provider.apiConfig.Contexts[params.Name]; exist {
				params.Name = fmt.Sprintf("%s.%s", params.Name, lang.RandomString(3))
			}
			context.Cluster = fmt.Sprintf("%s-cluster", params.Name)
			context.AuthInfo = fmt.Sprintf("%s-user", params.Name)

			if _, exist := provider.apiConfig.Clusters[context.Cluster]; exist {
				context.Cluster = fmt.Sprintf("%s-%s-cluster", params.Name, lang.RandomString(5))
			}
			if _, exist := provider.apiConfig.AuthInfos[context.AuthInfo]; exist {
				context.AuthInfo = fmt.Sprintf("%s-%s-user", params.Name, lang.RandomString(5))
			}
		} else {
			context.Cluster = fmt.Sprintf("%s-cluster", params.Name)
			context.AuthInfo = fmt.Sprintf("%s-user", params.Name)
		}

		// parsing - cluster
		cluster.Server = params.Cluster.Server
		if params.Cluster.CertificateAuthorityData != "" {
			ca, err := base64.StdEncoding.DecodeString(params.Cluster.CertificateAuthorityData)
			if err != nil {
				return fmt.Errorf("Unable to decode cerificate-authority-data (data=%s, cause=%s)", params.Cluster.CertificateAuthorityData, err)
			}
			cluster.CertificateAuthorityData = ca
		}

		// parsing - user
		if params.User.ClientCertificateData != "" {
			ca, err := base64.StdEncoding.DecodeString(params.User.ClientCertificateData)
			if err != nil {
				return fmt.Errorf("Unable to decode client-certificate-data (data=%s, cuase=%s)", params.User.ClientCertificateData, err)
			}
			user.ClientCertificateData = ca
		}

		if params.User.ClientKeyData != "" {
			ca, err := base64.StdEncoding.DecodeString(params.User.ClientKeyData)
			if err != nil {
				return fmt.Errorf("Unable to decode client-key-data (data=%s, cause=%s)", params.User.ClientKeyData, err)
			}
			user.ClientKeyData = ca
		}

		if params.User.Token != "" {
			user.Token = params.User.Token
		}

		provider.apiConfig.Clusters[context.Cluster] = cluster
		provider.apiConfig.AuthInfos[context.AuthInfo] = user
		provider.apiConfig.Contexts[params.Name] = context

		return provider.write()
	}

	clusters.RemoveCluster = func(name string) error {

		conf := provider.apiConfig.DeepCopy()

		if conf.Contexts[name] != nil {
			if conf.Clusters[conf.Contexts[name].Cluster] != nil {
				delete(conf.Clusters, conf.Contexts[name].Cluster)
			}
			if conf.AuthInfos[conf.Contexts[name].AuthInfo] != nil {
				delete(conf.AuthInfos, conf.Contexts[name].AuthInfo)
			}
			if conf.CurrentContext == name {
				conf.CurrentContext = ""
			}
			delete(conf.Contexts, name)

		} else {
			return fmt.Errorf("not found context %s", name)
		}

		provider.apiConfig = conf
		return provider.write()

	}

	clusters.GetClusterNames = func() []string {
		names := []string{}
		for k := range provider.apiConfig.Contexts {
			names = append(names, k)
		}
		return names
	}

	return clusters, nil

}

// ConfigMap update delay problem.
// database
//   - etcd?
//   - file base
//   - yaml/json base ? object base
//   - triggering
func getKubeConfigOptions(opts string) (*KubeConfigOptions, error) {

	options := &KubeConfigOptions{}
	// unmarshall kubeconfig
	if opts == "" || !strings.Contains(opts, "strategy=") {
		options.Strategy = KubeConfigStrategyFile
		options.Filename = opts
	} else {
		for _, e := range strings.Split(opts, ",") {
			parts := strings.Split(e, "=")
			if parts[0] == "strategy" {
				options.Strategy = KubeConfigStrategy(parts[1])
			} else if parts[0] == "configmap" {
				options.ConfigMap = parts[1]
			} else if parts[0] == "namespace" {
				options.Namespace = parts[1]
			} else if parts[0] == "filename" {
				options.Filename = parts[1]
			}
		}
	}

	if options.Strategy == "" {
		return nil, errors.New("KUBECONFIG strategy is empty")
	} else if !(options.Strategy == KubeConfigStrategyFile || options.Strategy == KubeConfigStrategyConfigmap) {
		return nil, errors.New(fmt.Sprintf("Not supported KUBECONFIG strategy (strategy=%s)", options.Strategy))
	} else if options.Strategy == KubeConfigStrategyConfigmap {
		if options.ConfigMap == "" {
			return nil, errors.New(fmt.Sprintf("ConfigMap name is empty for KUBECONFIG (strategy=%s)", options.Strategy))
		} else if options.Namespace == "" {
			options.Namespace = "default"
		} else if options.Filename == "" {
			options.Namespace = "config"
		}
	}

	return options, nil

}
