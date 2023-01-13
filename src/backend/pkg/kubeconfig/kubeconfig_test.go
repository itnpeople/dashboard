package kubeconfig

import (
	"context"
	"testing"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func TestKubeConfig(t *testing.T) {

	// File based KubeConfig
	t.Log("▶ File based KubeConfig")
	if conf, err := NewKubeContexts(""); err != nil {
		t.Error(err)
	} else {
		if client, err := conf.Client(conf.CurrentContext); err != nil {
			t.Error(err)
		} else {
			if err := printNamespaces(client, t); err != nil {
				t.Error(err)
			} else {
				t.Log("→ OK")
			}
		}
	}

	// ConfigMap based KubeConfig
	//  - prerequisite : in-cluster 환경 9http://itnp.kr/post/client-go)
	//  - kubectl create configmap kore-board-kubeconfig -n default --from-file=config=${HOME}/.kube/config
	//  - kubectl delete configmap kore-board-kubeconfig -n default
	t.Log("▶ ConfigMap based KubeConfig")
	if conf, err := NewKubeContexts("strategy=configmap,configmap=kore-board-kubeconfig,namespace=default,filename=config"); err != nil {
		t.Error(err)
	} else {
		if client, err := conf.Client(conf.CurrentContext); err != nil {
			t.Error(err)
		} else {
			if err := printNamespaces(client, t); err != nil {
				t.Error(err)
			} else {
				t.Log("→ OK")
			}
		}
	}

}

func printNamespaces(client *kubernetes.Clientset, t *testing.T) error {

	if list, err := client.CoreV1().Namespaces().List(context.TODO(), metaV1.ListOptions{}); err != nil {
		return err
	} else {
		t.Logf("  namespace (%d ea)", len(list.Items))
		for i, ns := range list.Items {
			t.Logf("    %d : %s", i+1, ns.ObjectMeta.Name)
		}
	}
	return nil
}
