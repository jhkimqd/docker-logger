package docker

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// CreateClient initializes a new Docker client.
func CreateClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return cli, nil
}

// InspectNetwork retrieves details about a Docker network.
func InspectNetwork(ctx context.Context, cli *client.Client, networkName string) (types.NetworkResource, error) {
	network, err := cli.NetworkInspect(ctx, networkName, types.NetworkInspectOptions{})
	if err != nil {
		return types.NetworkResource{}, err
	}
	return network, nil
}