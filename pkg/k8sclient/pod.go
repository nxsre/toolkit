package k8sclient

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"log"
)

func createPodWatcher(ctx context.Context, namespace, resName string) (watch.Interface, error) {
	//labelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", resName)
	//log.Printf("Creating watcher for POD with label: %s", labelSelector)

	opts := metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{},
		//LabelSelector: labelSelector,
		FieldSelector: fields.OneTermEqualSelector(metav1.ObjectNameField, resName).String(),
	}
	clientset, err := createClientSet()
	if err != nil {
		return nil, err
	}
	return clientset.CoreV1().Pods(namespace).Watch(ctx, opts)
}

func waitPodDeleted(ctx context.Context, namespace, resName string) error {
	watcher, err := createPodWatcher(ctx, namespace, resName)
	if err != nil {
		return err
	}

	defer watcher.Stop()

	for {
		select {
		case event := <-watcher.ResultChan():
			if event.Type == watch.Deleted {
				log.Printf("The POD \"%s\" is deleted", resName)
				return nil
			}

		case <-ctx.Done():
			log.Printf("Exit from waitPodDeleted for POD \"%s\" because the context is done", resName)
			return nil
		}
	}
}

func waitPodRunning(ctx context.Context, namespace, resName string) error {
	watcher, err := createPodWatcher(ctx, namespace, resName)
	if err != nil {
		return err
	}

	defer watcher.Stop()

	for {
		select {
		case event := <-watcher.ResultChan():
			pod := event.Object.(*v1.Pod)

			if pod.Status.Phase == v1.PodRunning {
				log.Printf("The POD \"%s\" is running", resName)
				return nil
			}
		case <-ctx.Done():
			log.Printf("Exit from waitPodRunning for POD \"%s\" because the context is done", resName)
			return nil
		}
	}
}
