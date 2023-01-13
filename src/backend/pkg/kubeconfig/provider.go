package kubeconfig

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

	provider := &KubeConfigProvider{KubeConfigOptions: options}
	if provider.Strategy == KubeConfigStrategyConfigmap {

		// validate (configmap-mode is not supported on in-cluster)
		rest, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		clientset, err := kubernetes.NewForConfig(rest)
		if err != nil {
			return nil, err
		}

		// read from configmap
		var resoureVersion string
		provider.Read = func() error {

			clientset, err := kubernetes.NewForConfig(rest)
			if err != nil {
				return err
			}

			cm, err := clientset.CoreV1().ConfigMaps(provider.Namespace).Get(context.TODO(), provider.ConfigMap, v1.GetOptions{})
			if err != nil {
				return err
			}
			resoureVersion = cm.ObjectMeta.ResourceVersion

			if cm.Data[provider.Filename] == "" {
				return fmt.Errorf("kubeconfig data is empty namespace=%s, configmap=%s, filename=%s", provider.Namespace, provider.ConfigMap, provider.Filename)
			}
			clientConfig, err := clientcmd.NewClientConfigFromBytes([]byte(cm.Data[provider.Filename]))
			if err != nil {
				return err
			}

			//apiConfig := &api.Config{}
			apiConfig, err := clientConfig.RawConfig()
			if err != nil {
				return err
			}

			provider.APIConfig = apiConfig.DeepCopy()

			return nil
		}

		// write to configmap
		provider.Write = func() error {

			cm, err := clientset.CoreV1().ConfigMaps(provider.Namespace).Get(context.TODO(), provider.ConfigMap, v1.GetOptions{})
			if err != nil {
				return err
			}

			b, err := clientcmd.Write(*provider.APIConfig)
			if err != nil {
				return err
			}
			cm.Data[provider.Filename] = string(b)

			_, err = clientset.CoreV1().ConfigMaps(provider.Namespace).Update(context.TODO(), cm, v1.UpdateOptions{})
			if err != nil {
				return err
			}
			return nil
		}

		// is modified
		provider.IsModified = func() bool {
			if clientset, err := kubernetes.NewForConfig(rest); err == nil {
				if cm, err := clientset.CoreV1().ConfigMaps(provider.Namespace).Get(context.TODO(), provider.ConfigMap, v1.GetOptions{}); err == nil {
					return resoureVersion != cm.ObjectMeta.ResourceVersion
				}
			}
			return false
		}

	} else {

		// read from file
		var modifiedTime int64
		provider.Read = func() error {
			var configLoadingRules clientcmd.ClientConfigLoader
			if options.Filename == "" {
				configLoadingRules = clientcmd.NewDefaultClientConfigLoadingRules()
			} else {
				configLoadingRules = &clientcmd.ClientConfigLoadingRules{ExplicitPath: provider.Filename}
			}
			apiConfig, err := configLoadingRules.Load()
			provider.Filename = configLoadingRules.GetDefaultFilename()
			if err != nil {
				return err
			}
			provider.APIConfig = apiConfig.DeepCopy()

			file, err := os.Stat(provider.Filename)
			if err != nil {
				return err
			}
			modifiedTime = file.ModTime().UnixNano()
			return nil
		}

		// write to file
		provider.Write = func() error {
			if err := clientcmd.WriteToFile(*provider.APIConfig, provider.Filename); err != nil {
				return err
			}
			return nil
		}

		// is modified
		provider.IsModified = func() bool {
			file, err := os.Stat(provider.Filename)
			if err != nil {
				return true
			}
			return modifiedTime != file.ModTime().UnixNano()
		}

	}

	return provider, nil

}
