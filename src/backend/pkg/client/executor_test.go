package client

import (
	"context"
	"strings"
	"testing"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/dynamic"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var busyboxYaml = `apiVersion: v1
kind: Pod
metadata:
  name: busybox
  labels:
    app: koreboard-backend-client-test-busybox
spec:
  containers:
  - name: busybox
    image: busybox
    command:
      - sleep
      - "10"
    imagePullPolicy: IfNotPresent
  restartPolicy: Always`

var namespaceYaml = `apiVersion: v1
kind: Namespace
metadata:
  name: default-test`

func TestKubeExecutorNamespace(t *testing.T) {
	var restConfig *rest.Config
	var err error
	if restConfig, err = getRestConfig(); err != nil {
		t.Error(err)
		return
	}

	t.Log("▶ POST() - Namespace")
	if r, err := NewKubeExecutor(restConfig).POST(strings.NewReader(namespaceYaml), false); err != nil {
		t.Error(err)
	} else {
		t.Logf("→ OK (name=%s, namespace=%s, resourceVersion=%s)", r.GetName(), r.GetNamespace(), r.GetResourceVersion())
	}

	t.Log("▶ Delete()")
	var client *dynamic.DynamicClient
	if client, err = dynamic.NewForConfig(restConfig); err != nil {
		t.Error(err)
	} else {
		gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}
		if err := client.Resource(gvr).Delete(context.TODO(), "default-test", metaV1.DeleteOptions{}); err != nil {
			t.Error(err)
		} else {
			t.Logf("→ OK")
		}
	}

}

func TestKubeExecutorPod(t *testing.T) {

	var restConfig *rest.Config
	var err error
	if restConfig, err = getRestConfig(); err != nil {
		t.Error(err)
		return
	}

	t.Log("▶ POST() - Pod")
	if r, err := NewKubeExecutor(restConfig).Namespace("default").POST(strings.NewReader(busyboxYaml), false); err != nil {
		t.Error(err)
	} else {
		t.Logf("→ OK (name=%s, namespace=%s, resourceVersion=%s)", r.GetName(), r.GetNamespace(), r.GetResourceVersion())
	}

	t.Log("▶ Watch()")
	var client *dynamic.DynamicClient
	if client, err = dynamic.NewForConfig(restConfig); err != nil {
		t.Error(err)
		return
	}
	// LabelSelector: "app=koreboard-backend-client-test-busybox"
	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	opts := metaV1.SingleObject(metaV1.ObjectMeta{Name: "busybox"})
	if watcher, err := client.Resource(gvr).Namespace("default").Watch(context.TODO(), opts); err != nil {
		t.Error(err)
	} else {
		timer := time.NewTimer(30 * time.Second)
		defer watcher.Stop()
	loop:
		for {
			select {
			case e := <-watcher.ResultChan():
				if e.Object == nil {
					t.Errorf("Fail to watch pod")
					break loop
				} else {
					a, _ := e.Object.(*unstructured.Unstructured)
					pod := &v1.Pod{}
					if err := runtime.DefaultUnstructuredConverter.FromUnstructured(a.UnstructuredContent(), pod); err != nil {
						t.Error(err)
					} else {
						if pod.GetName() == "busybox" {
							for _, c := range pod.Status.Conditions {
								if c.Type == v1.PodReady {
									t.Logf("→ PodReady")
									break loop
								}
							}
							break loop
						}
					}
				}
			case <-timer.C:
				t.Error("timeout")
				break loop
			}
		}
	}
	t.Logf("→ OK")

	t.Log("▶ Delete()")
	if err := client.Resource(gvr).Namespace("default").Delete(context.TODO(), "busybox", metaV1.DeleteOptions{}); err != nil {
		t.Error(err)
	} else {
		t.Logf("→ OK")
	}

}

func getRestConfig() (*rest.Config, error) {

	if apiConfig, err := clientcmd.NewDefaultClientConfigLoadingRules().Load(); err != nil {
		return nil, err
	} else if restConfig, err := clientcmd.NewNonInteractiveClientConfig(*apiConfig, apiConfig.CurrentContext, &clientcmd.ConfigOverrides{}, nil).ClientConfig(); err != nil {
		return nil, err
	} else {
		return restConfig, nil
	}

}
