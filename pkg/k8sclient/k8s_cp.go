package k8sclient

import (
	"bytes"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/kubernetes/scheme"
)

func KubeCopy(namespace, src, dest string) error {
	restCfg, err := createRestConfig()
	if err != nil {
		return err
	}
	clientset, err := createClientSet()
	if err != nil {
		return err
	}
	restCfg.GroupVersion = &schema.GroupVersion{Version: "v1"}
	restCfg.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	restCfg.APIPath = "/api"

	in, out, errOut := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
	defer func() {
		in.Reset()
		out.Reset()
		errOut.Reset()
	}()
	cpOption := NewCopyOptions(genericiooptions.IOStreams{
		In:     in,
		Out:    out,
		ErrOut: errOut,
	})

	cpOption.Namespace = namespace
	cpOption.Clientset = clientset
	cpOption.ClientConfig = restCfg
	return cpOption.Run(src, dest)
}
