package clusters

import (
	"context"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewKubeConfigProvider(options *KubeConfigOptions) (*KubeConfigProvider, error) {

	provider := &KubeConfigProvider{options: options}
	if options.Strategy == KubeConfigStrategyConfigmap {

		// validate (configmap-mode is not supported on in-cluster)
		inClusterConfig, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}

		// read from configmap
		var resoureVersion string
		provider.read = func() error {

			clientset, err := kubernetes.NewForConfig(inClusterConfig)
			if err != nil {
				return err
			}

			cm, err := clientset.CoreV1().ConfigMaps(options.Namespace).Get(context.TODO(), options.ConfigMap, v1.GetOptions{})
			if err != nil {
				return err
			}
			resoureVersion = cm.ObjectMeta.ResourceVersion

			if cm.Data[options.Filename] == "" {
				return fmt.Errorf("kubeconfig data is empty namespace=%s, configmap=%s, filename=%s", options.Namespace, options.ConfigMap, options.Filename)
			}
			clientConfig, err := clientcmd.NewClientConfigFromBytes([]byte(cm.Data[options.Filename]))
			if err != nil {
				return err
			}

			//apiConfig := &api.Config{}
			apiConfig, err := clientConfig.RawConfig()
			if err != nil {
				return err
			}

			provider.apiConfig = apiConfig.DeepCopy()

			return nil
		}

		// write to configmap
		provider.write = func() error {

			clientset, err := kubernetes.NewForConfig(inClusterConfig)
			if err != nil {
				return err
			}

			cm, err := clientset.CoreV1().ConfigMaps(options.Namespace).Get(context.TODO(), options.ConfigMap, v1.GetOptions{})
			if err != nil {
				return err
			}

			b, err := clientcmd.Write(*provider.apiConfig)
			if err != nil {
				return err
			}
			cm.Data[options.Filename] = string(b)

			_, err = clientset.CoreV1().ConfigMaps(options.Namespace).Update(context.TODO(), cm, v1.UpdateOptions{})
			if err != nil {
				return err
			}
			return nil
		}

		// is modified
		provider.isModified = func() bool {
			if clientset, err := kubernetes.NewForConfig(inClusterConfig); err == nil {
				if cm, err := clientset.CoreV1().ConfigMaps(options.Namespace).Get(context.TODO(), options.ConfigMap, v1.GetOptions{}); err == nil {
					return resoureVersion != cm.ObjectMeta.ResourceVersion
				}
			}
			return false
		}

	} else {

		// read from file
		var modifiedTime int64
		provider.read = func() error {
			var configLoadingRules clientcmd.ClientConfigLoader
			if options.Filename == "" {
				configLoadingRules = clientcmd.NewDefaultClientConfigLoadingRules()
			} else {
				configLoadingRules = &clientcmd.ClientConfigLoadingRules{ExplicitPath: options.Filename}
			}
			apiConfig, err := configLoadingRules.Load()
			options.Filename = configLoadingRules.GetDefaultFilename()
			if err != nil {
				return err
			}
			provider.apiConfig = apiConfig.DeepCopy()

			file, err := os.Stat(options.Filename)
			if err != nil {
				return err
			}
			modifiedTime = file.ModTime().UnixNano()
			return nil
		}

		// write to file
		provider.write = func() error {
			if err := clientcmd.WriteToFile(*provider.apiConfig, options.Filename); err != nil {
				return err
			}
			return nil
		}

		// is modified
		provider.isModified = func() bool {
			file, err := os.Stat(options.Filename)
			if err != nil {
				return true
			}
			return modifiedTime != file.ModTime().UnixNano()
		}

	}

	return provider, nil

}
