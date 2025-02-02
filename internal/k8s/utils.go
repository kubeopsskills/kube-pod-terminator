package k8s

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

// GetConfig gets parameters to generate rest.Config and returns it
func GetConfig(kubeConfigPath string, inCluster bool) (*rest.Config, error) {
	var (
		config *rest.Config
		err    error
	)

	if inCluster {
		if config, err = rest.InClusterConfig(); err != nil {
			return nil, err
		}
	} else {
		if config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// GetClientSet generates and returns k8s.Clientset using rest.Config
func GetClientSet(config *rest.Config) (*kubernetes.Clientset, error) {
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientSet, nil
}

func getTerminatingPods(ctx context.Context, clientSet kubernetes.Interface, namespace string) ([]v1.Pod, error) {
	var (
		resultSlice []v1.Pod
		pods        *v1.PodList
		err         error
	)

	if pods, err = clientSet.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{}); err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		deletionTimestamp := pod.ObjectMeta.DeletionTimestamp
		if deletionTimestamp != nil && deletionTimestamp.Add(time.Duration(opts.TerminatingStateMinutes)*time.Minute).Before(time.Now()) {
			resultSlice = append(resultSlice, pod)
		}
	}

	return resultSlice, nil
}

func getEvictedPods(ctx context.Context, clientSet kubernetes.Interface, namespace string) ([]v1.Pod, error) {
	var (
		evictedPods []v1.Pod
		pods        *v1.PodList
		err         error
	)

	if pods, err = clientSet.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{}); err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		if pod.Status.Reason == "Evicted" {
			evictedPods = append(evictedPods, pod)
		}
	}

	return evictedPods, nil
}
