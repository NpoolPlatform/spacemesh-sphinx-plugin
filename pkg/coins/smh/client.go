package smh

import (
	"context"
	"time"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/endpoints"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/utils"
	"github.com/NpoolSpacemesh/spacemesh-plugin/client"
)

const (
	MinNodeNum       = 1
	MaxRetries       = 3
	retriesSleepTime = 200 * time.Millisecond
	reqTimeout       = 3 * time.Second
)

type SClientI interface {
	GetNode(ctx context.Context, endpointmgr *endpoints.Manager) (*client.Client, error)
	WithClient(ctx context.Context, fn func(context.Context, *client.Client) (bool, error)) error
}

type SClients struct{}

func (sClients SClients) GetNode(ctx context.Context, endpointmgr *endpoints.Manager) (*client.Client, error) {
	endpoint, err := endpointmgr.Peek()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, reqTimeout)
	defer cancel()

	client := client.NewClient(endpoint, false)
	err = client.Connect()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (sClients *SClients) WithClient(ctx context.Context, fn func(ctx context.Context, c *client.Client) (bool, error)) error {
	var (
		apiErr, err error
		retry       bool
		client      *client.Client
	)
	endpointmgr, err := endpoints.NewManager()
	if err != nil {
		return err
	}

	for i := 0; i < utils.MinInt(MaxRetries, endpointmgr.Len()); i++ {
		if i > 0 {
			time.Sleep(retriesSleepTime)
		}

		client, err = sClients.GetNode(ctx, endpointmgr)
		if err != nil {
			continue
		}

		retry, apiErr = fn(ctx, client)
		if !retry {
			return apiErr
		}
	}
	if apiErr != nil {
		return apiErr
	}
	return err
}

func Client() SClientI {
	return &SClients{}
}
