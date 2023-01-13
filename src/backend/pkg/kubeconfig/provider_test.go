package kubeconfig

import (
	"context"
	"os/exec"
	"strconv"
	"testing"
	"time"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func TestFileKubeConfigProvider(t *testing.T) {

	var provider *KubeConfigProvider
	var options *KubeConfigOptions
	var err error

	// create a provider
	if options, err = getKubeConfigOptions(""); err != nil {
		t.Error(err)
	} else {
		if provider, err = NewKubeConfigProvider(options); err != nil {
			t.Error(err)
		} else {
			if err := provider.Read(); err != nil {
				t.Error(err)
			}
		}
	}
	t.Log("▶ Create a provider → OK")

	// vlidate modified config-file
	if provider.IsModified() {
		t.Error("Fail to run IsModified()")
	} else {
		cmd := exec.Command("touch", provider.Filename)
		if err = cmd.Run(); err != nil {
			t.Error(err)
			return
		}
		if !provider.IsModified() {
			t.Error("The file is not modified but it have to modified ")
		} else {
			t.Log("▶ Validate modified config-file → OK")
		}
	}

}

// Rrerequisite :
//   - in-cluster 환경 구축
//   - http://itnp.kr/post/client-go
//
// kubectl create configmap kore-board-kubeconfig -n default --from-file=config=${HOME}/.kube/config
// kubectl delete configmap kore-board-kubeconfig -n default
func TestConfigMapKubeConfigProvider(t *testing.T) {

	var provider *KubeConfigProvider
	var options *KubeConfigOptions
	var err error

	if options, err = getKubeConfigOptions("strategy=configmap,configmap=kore-board-kubeconfig,namespace=default,filename=config"); err != nil {
		t.Error(err)
	} else {
		if provider, err = NewKubeConfigProvider(options); err != nil {
			t.Error(err)
		} else {
			if err = provider.Read(); err != nil {
				t.Error(err)
			}
		}
	}
	t.Log("▶ Create a provider → OK")

	// vlidate modified config-file
	if provider.IsModified() {
		t.Error("Fail to run IsModified()")
	} else {
		if rest, err := rest.InClusterConfig(); err != nil {
			t.Error(err)
		} else if clientset, err := kubernetes.NewForConfig(rest); err != nil {
			t.Error(err)
		} else if cm, err := clientset.CoreV1().ConfigMaps(provider.Namespace).Get(context.TODO(), provider.ConfigMap, metaV1.GetOptions{}); err != nil {
			t.Error(err)
		} else {
			//update configmap
			cm.Data["modifiy.dat"] = strconv.FormatInt(time.Now().UnixNano(), 10)
			if _, err = clientset.CoreV1().ConfigMaps(provider.Namespace).Update(context.TODO(), cm, metaV1.UpdateOptions{}); err != nil {
				t.Error(err)
				return
			}
			if !provider.IsModified() {
				t.Error("The configmap is not modified but it have to modified ")
			} else {
				t.Log("▶ Validate modified configmap → OK")
			}

		}

	}

}
