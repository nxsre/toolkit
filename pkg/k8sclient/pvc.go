package k8sclient

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"log"
)

func createPvcWatcher(ctx context.Context, namespace, resName string) (watch.Interface, error) {
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
	return clientset.CoreV1().PersistentVolumeClaims(namespace).Watch(ctx, opts)
}

func waitPvcDeleted(ctx context.Context, namespace, resName string) error {
	watcher, err := createPvcWatcher(ctx, namespace, resName)
	if err != nil {
		return err
	}

	defer watcher.Stop()

	for {
		select {
		case event := <-watcher.ResultChan():
			if event.Type == watch.Deleted {
				log.Printf("The PVC \"%s\" is deleted", resName)
				return nil
			}

		case <-ctx.Done():
			log.Printf("Exit from waitPvcDeleted for PVC \"%s\" because the context is done", resName)
			return nil
		}
	}
}

func waitPvcBound(ctx context.Context, namespace, resName string) error {
	watcher, err := createPvcWatcher(ctx, namespace, resName)
	if err != nil {
		return err
	}
	defer watcher.Stop()
	for {
		select {
		case event := <-watcher.ResultChan():
			pvc := event.Object.(*v1.PersistentVolumeClaim)

			if pvc.Status.Phase == v1.ClaimBound {
				log.Printf("The PVC \"%s\" is bound", resName)
				return nil
			}
		case <-ctx.Done():
			log.Printf("Exit from waitPvcRunning for PVC \"%s\" because the context is done", resName)
			return nil
		}
	}
}
