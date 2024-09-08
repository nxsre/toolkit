package k8sclient

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/pager"
	"k8s.io/utils/trace"
	"time"
)

func TimeNewFilteredPodInformer() ([]*v1.Pod, error) {
	clientset, _ := ClientSet()
	// 不分页
	//options := metav1.ListOptions{ResourceVersion: "0"}
	// 分页
	options := metav1.ListOptions{Limit: 200}

	initTrace := trace.New("Reflector ListAndWatch", trace.Field{Key: "name", Value: "pod"})
	defer initTrace.LogIfLong(1 * time.Millisecond)
	var list runtime.Object
	var paginatedResult bool
	var err error
	listCh := make(chan struct{}, 1)
	panicCh := make(chan interface{}, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				panicCh <- r
			}
		}()
		// Attempt to gather list in chunks, if supported by listerWatcher, if not, the first
		// list request will return the full response.
		pager := pager.New(pager.SimplePageFunc(func(opts metav1.ListOptions) (runtime.Object, error) {
			lw := &cache.ListWatch{
				ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
					return clientset.CoreV1().Pods(v1.NamespaceAll).List(context.TODO(), options)
				},
				WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
					return clientset.CoreV1().Pods(v1.NamespaceAll).Watch(context.TODO(), options)
				},
			}
			return lw.List(opts)
		}))

		list, paginatedResult, err = pager.List(context.Background(), options)
		initTrace.Step("Objects listed: ")
		_ = paginatedResult
		//fmt.Println("list END, is pager ", paginatedResult)
		if err != nil {
			fmt.Println("error is : ", err.Error())
		}
		close(listCh)
	}()
	select {
	case r := <-panicCh:
		panic(r)
	case <-listCh:
	}

	initTrace.Step("Resource version extracted")
	items, err := meta.ExtractList(list)
	//fmt.Println("list items size is : ", len(items))
	if err != nil {
		return nil, fmt.Errorf("unable to understand list result %#v (%v)", list, err)
	}
	pods := []*v1.Pod{}
	for _, n := range items {
		pods = append(pods, n.(*v1.Pod))
	}
	initTrace.Step("Objects extracted")
	return pods, nil
}

func TimeNewIndexerInformer() ([]*v1.Pod, error) {
	clientset, _ := ClientSet()
	options := metav1.ListOptions{ResourceVersion: "0"}

	initTrace := trace.New("Reflector ListAndWatch", trace.Field{Key: "name", Value: "pod"})
	defer initTrace.LogIfLong(1 * time.Millisecond)
	var list runtime.Object
	var paginatedResult bool
	var err error
	listCh := make(chan struct{}, 1)
	panicCh := make(chan interface{}, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				panicCh <- r
			}
		}()
		// Attempt to gather list in chunks, if supported by listerWatcher, if not, the first
		// list request will return the full response.
		pager := pager.New(pager.SimplePageFunc(func(opts metav1.ListOptions) (runtime.Object, error) {
			lw := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", v1.NamespaceAll, fields.Everything())
			return lw.List(opts)
		}))

		list, paginatedResult, err = pager.List(context.Background(), options)
		initTrace.Step("Objects listed: ")
		fmt.Println("list END, is pager ", paginatedResult)
		if err != nil {
			fmt.Println("error is : ", err.Error())
		}
		close(listCh)
	}()
	select {
	case r := <-panicCh:
		panic(r)
	case <-listCh:
	}

	initTrace.Step("Resource version extracted")
	items, err := meta.ExtractList(list)
	fmt.Println("list items size is : ", len(items))
	if err != nil {
		return nil, fmt.Errorf("unable to understand list result %#v (%v)", list, err)
	}
	pods := []*v1.Pod{}
	for _, n := range items {
		pods = append(pods, n.(*v1.Pod))
	}
	initTrace.Step("Objects extracted")
	return pods, nil
}
