package container

import (
	"context"
	"github.com/containerd/containerd/v2/client"
	"github.com/containerd/nerdctl/v2/pkg/clientutil"
	"github.com/containerd/nerdctl/v2/pkg/idutil/containerwalker"
	"log"
	// container 是 nerdctl 封装好的所有方法，可以参考
	//"github.com/containerd/nerdctl/v2/pkg/cmd/container"
)

var DefaultEndpoint = "/run/containerd/containerd.sock"

func ContainerWalker(ctx context.Context, client *client.Client, req string, found containerwalker.OnFound) (n int, err error) {
	containerwalker := &containerwalker.ContainerWalker{
		Client:  client,
		OnFound: found,
	}

	return containerwalker.Walk(ctx, req)
}

func NewK8sContainerClient(ctx context.Context) (*client.Client, context.Context, context.CancelFunc, error) {
	client, ctx, cancel, err := clientutil.NewClient(ctx, "k8s.io", DefaultEndpoint)
	if err != nil {
		log.Println(err)
	}
	return client, ctx, cancel, err
}
