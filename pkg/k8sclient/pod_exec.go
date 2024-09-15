package k8sclient

import (
	"bytes"
	"context"
	"github.com/nxsre/toolkit/pkg/k8sclient/spdy"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"time"

	"net/http"
	"strings"
)

func ExecInPod(ctx context.Context, namespace, podName, command, containerName string) (string, string, error) {
	restConnfig, err := createRestConfig()
	if err != nil {
		return "", "", err
	}
	//k8sCli, err := createClientSet()
	//if err != nil {
	//	return "", "", err
	//}
	cmd := []string{
		"sh",
		"-c",
		command,
	}
	const tty = false

	// 这里只是生成一个请求，不会调用 Do 方法，所以 restConnfig 的 transport 在这里不生效
	client, err := corev1.NewForConfig(restConnfig)
	req := client.RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).SubResource("exec").Param("container", containerName)

	req.VersionedParams(
		&v1.PodExecOptions{
			Command: cmd,
			Stdin:   false,
			Stdout:  true,
			Stderr:  true,
			TTY:     tty,
		},
		scheme.ParameterCodec,
	)

	var stdout, stderr bytes.Buffer

	tlsConfig, err := rest.TLSConfigFor(restConnfig)
	if err != nil {
		return "", "", err
	}
	proxy := http.ProxyFromEnvironment
	if restConnfig.Proxy != nil {
		proxy = restConnfig.Proxy
	}

	upgradeRoundTripper := spdy.NewRoundTripperWithConfig(spdy.RoundTripperConfig{
		TLS:         tlsConfig,
		Proxier:     proxy,
		PingPeriod:  time.Second * 5,
		DialContext: retryDial,
	})

	wrapper, err := rest.HTTPWrappersForConfig(restConnfig, upgradeRoundTripper)
	if err != nil {
		return "", "", err
	}

	exec, err := remotecommand.NewSPDYExecutorForTransports(wrapper, upgradeRoundTripper, "POST", req.URL())
	if err != nil {
		return "", "", err
	}
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return "", "", err
	}
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}
