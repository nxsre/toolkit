package k8sclient

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
)

func SendEvent(object runtime.Object, eventtype, reason, message string) {
	executable, err := os.Executable()
	if err != nil {
		panic(err)
	}
	processName := filepath.Base(executable)

	clientset, err := createClientSet()
	if err != nil {
		return
	}
	// 创建 eventRecorder, 用于发送 event 到 k8s
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: clientset.CoreV1().Events("")})
	eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: processName}).Event(
		object, eventtype, reason, message)

}
