package container

import (
	"context"
	"github.com/containerd/containerd/v2/client"
	"github.com/containerd/nerdctl/v2/pkg/idutil/containerwalker"
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

func NewClient() {

}
