package k8sclient

import (
	"context"
	"github.com/containerd/containerd"
	"github.com/containerd/nerdctl/pkg/clientutil"
	"github.com/containerd/nerdctl/pkg/idutil/containerwalker"
	"log"
	// container 是 nerdctl 封装好的所有方法，可以参考
	//"github.com/containerd/nerdctl/v2/pkg/cmd/container"
)

var DefaultEndpoint = "/run/containerd/containerd.sock"

func ContainerWalker(ctx context.Context, client *containerd.Client, req string, found containerwalker.OnFound) (n int, err error) {
	containerwalker := &containerwalker.ContainerWalker{
		Client:  client,
		OnFound: found,
	}

	return containerwalker.Walk(ctx, req)
}

func NewContainerClient(ctx context.Context, endpoint string) (*containerd.Client, context.Context, context.CancelFunc, error) {
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	log.Println("connect to ", endpoint)
	client, ctx, cancel, err := clientutil.NewClient(ctx, "k8s.io", endpoint)
	if err != nil {
		log.Println(err)
	}
	return client, ctx, cancel, err
}
