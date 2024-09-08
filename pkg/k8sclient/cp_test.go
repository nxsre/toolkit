package k8sclient

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"log/slog"
)

func main() {
	restConfig, err := createRestConfig()
	if err != nil {
		slog.Error("初始化客户端失败")
	}

	restConfig.GroupVersion = &schema.GroupVersion{Version: "v1"}
	restConfig.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	restConfig.APIPath = "/api"

	clientset, _ := createClientSet()

	KubeCopy(clientset, restConfig, "default", "./daemonset.yaml",
		fmt.Sprintf("%v:%v", "cph-test-20240119-004", "/data/local/tmp/cphnative-server-current.tgz"))
}
