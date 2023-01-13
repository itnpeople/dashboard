package kubeconfig

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

func NewKubeContexts(opts string) (*KubeContexts, error) {

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
			if err = provider.Read(); err != nil {
				return nil, err
			}
		}
	}

	//
	conf := &KubeContexts{}
	conf.RESTConfigs = make(map[string]*rest.Config)
	if apiConfig := provider.APIConfig; apiConfig == nil {
		if inClusterConfig, err := rest.InClusterConfig(); err == nil {
			conf.RESTConfigs[IN_CLUSTER_NAME] = inClusterConfig
			conf.CurrentContext = IN_CLUSTER_NAME
		}
	} else {
		conf.CurrentContext = apiConfig.CurrentContext
		for ctx := range apiConfig.Contexts {
			if restConfig, err := clientcmd.NewNonInteractiveClientConfig(*apiConfig, ctx, &clientcmd.ConfigOverrides{}, nil).ClientConfig(); err == nil {
				conf.RESTConfigs[ctx] = restConfig
				if conf.CurrentContext == "" {
					conf.CurrentContext = ctx
				}
			}
		}
	}

	// kubernetes.Clientset
	conf.Client = func(ctx string) (*kubernetes.Clientset, error) {

		if provider.IsModified() {
			if err := provider.Read(); err != nil {
				return nil, err
			}
		}
		if restConfig, ok := conf.RESTConfigs[ctx]; ok {
			return kubernetes.NewForConfig(restConfig)
		} else {
			return nil, errors.New(fmt.Sprintf("The Kubernetes-Client could not be created because the context was not found. (context=%s)", ctx))
		}
	}

	// versioned.Clientset
	conf.VersionedClient = func(ctx string) (*versioned.Clientset, error) {
		if provider.IsModified() {
			if err := provider.Read(); err != nil {
				return nil, err
			}
		}
		if restConfig, ok := conf.RESTConfigs[ctx]; ok {
			return versioned.NewForConfig(restConfig)
		} else {
			return nil, errors.New(fmt.Sprintf("The Versioned-Client could not be created because the context was not found. (context=%s)", ctx))
		}

	}

	// dynamic.DynamicClient
	conf.DynamicClient = func(ctx string) (*dynamic.DynamicClient, error) {
		if provider.IsModified() {
			if err := provider.Read(); err != nil {
				return nil, err
			}
		}
		if restConfig, ok := conf.RESTConfigs[ctx]; ok {
			return dynamic.NewForConfig(restConfig)
		} else {
			return nil, errors.New(fmt.Sprintf("The Dynamic-Client could not be created because the context was not found. (context=%s)", ctx))
		}
	}

	return conf, nil

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
				options.Strategy = parts[1]
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
