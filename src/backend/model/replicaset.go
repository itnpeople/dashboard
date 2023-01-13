package model

// base code : https://github.com/kubernetes/dashboard/tree/master/src/app/backend/resource/replicaset

import (
	"context"

	appsV1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	"github.com/kore3lab/dashboard/backend/pkg/lang"
)

// returns a subset of pods controlled by given deployment.
func GetReplicaSetPods(apiClient *kubernetes.Clientset, namespace string, name string) ([]v1.Pod, *v1.PodSpec, error) {

	replicaset, err := apiClient.AppsV1().ReplicaSets(namespace).Get(context.TODO(), name, metaV1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}

	labelSelector := labels.SelectorFromSet(replicaset.Spec.Selector.MatchLabels)

	podList, err := GetPodsMatchLabels(apiClient, namespace, labelSelector)
	if err != nil {
		return nil, nil, err
	}
	return lang.FilterPodsByControllerRef(replicaset, podList.Items), &replicaset.Spec.Template.Spec, nil

}

// return a subset of replicasets by given labelSelector
func GetReplicaSetMatchLabels(k8sClient *kubernetes.Clientset, namespace string, labelSelector labels.Selector) (*appsV1.ReplicaSetList, error) {

	rsList, err := k8sClient.AppsV1().ReplicaSets(namespace).List(context.TODO(), metaV1.ListOptions{LabelSelector: labelSelector.String()})
	if err != nil {
		return nil, err
	}

	return rsList, nil

}

// replicaset's available-ready count in a cluster
func GetReplicaSetsReady(apiClient *kubernetes.Clientset, options metaV1.ListOptions) (available int, ready int, err error) {

	list, err := apiClient.AppsV1().ReplicaSets("").List(context.TODO(), options)
	if err != nil {
		return available, ready, err
	}
	available = len(list.Items)
	for _, m := range list.Items {
		if m.Status.Replicas == m.Status.ReadyReplicas {
			ready += 1
		}
	}
	return available, ready, err

}
