/*
*
"k8s.io/client-go/dynamic"

	관련소스 : https://github.com/kubernetes/client-go/tree/master/dynamic

Kubernetes API Concepts

	https://kubernetes.io/docs/reference/using-api/api-concepts/

Patch

	https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/

활용예제 참조

	kubernetes-dashboard
	    https://github.com/kubernetes/dashboard/blob/master/src/app/backend/resource/deployment/deploy.go
	    DeployAppFromFile() 함수
*/
package client

import (
	"context"
	"fmt"
	"io"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// // RestfulClient 리턴
func (executor *KubeExecutor) Namespace(namespace string) *KubeExecutor {
	executor.namespace = namespace
	return executor
}

func NewKubeExecutor(config *rest.Config) *KubeExecutor {

	executor := &KubeExecutor{}

	executor.POST = func(payload io.Reader, isUpdate bool) (output *unstructured.Unstructured, err error) {

		d := yaml.NewYAMLOrJSONDecoder(payload, 4096)
		for {
			// payload 읽기
			data := &unstructured.Unstructured{}
			if err = d.Decode(data); err != nil {
				if err == io.EOF {
					return output, err
				}
				return output, err
			}

			// version kind
			version := data.GetAPIVersion()
			kind := data.GetKind()
			gv, err := schema.ParseGroupVersion(version)
			if err != nil {
				gv = schema.GroupVersion{Version: version}
			}

			// api version 에 해당하는 resource 정보 조회
			discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
			if err != nil {
				return output, err
			}

			apiResourceList, err := discoveryClient.ServerResourcesForGroupVersion(version)
			if err != nil {
				return output, err
			}

			var resource *v1.APIResource
			for _, apiResource := range apiResourceList.APIResources {
				if apiResource.Kind == kind && !strings.Contains(apiResource.Name, "/") {
					resource = &apiResource
					break
				}
			}
			if resource == nil {
				err = fmt.Errorf("unknown resource kind: %s", kind)
				return output, err
			}
			gvr := schema.GroupVersionResource{Group: gv.Group, Version: gv.Version, Resource: resource.Name}

			// 실행
			dynamicClient, err := dynamic.NewForConfig(config)
			if err != nil {
				return output, err
			}

			if resource.Namespaced && executor.namespace == "" {
				executor.namespace = data.GetNamespace()
			}

			// update 인 경우 resourceVersion 을 조회 & 수정
			if isUpdate {
				r, err := dynamicClient.Resource(gvr).Namespace(executor.namespace).Get(context.TODO(), data.GetName(), v1.GetOptions{})
				if err != nil {
					return output, err
				}
				data.SetResourceVersion(r.GetResourceVersion())
				if resource.Namespaced {
					output, err = dynamicClient.Resource(gvr).Namespace(executor.namespace).Update(context.TODO(), data, v1.UpdateOptions{})
				} else {
					output, err = dynamicClient.Resource(gvr).Update(context.TODO(), data, v1.UpdateOptions{})
				}
			} else {
				if resource.Namespaced {
					output, err = dynamicClient.Resource(gvr).Namespace(executor.namespace).Create(context.TODO(), data, v1.CreateOptions{})
				} else {
					output, err = dynamicClient.Resource(gvr).Create(context.TODO(), data, v1.CreateOptions{})
				}
			}

			return output, err

		}

	}

	return executor

}
