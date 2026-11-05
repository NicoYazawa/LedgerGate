//go:build integration

package testkit

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const natsImage = "nats:2.10-alpine"

type NATS struct {
	container testcontainers.Container
	url       string
}

func StartNATS(ctx context.Context) (*NATS, error) {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        natsImage,
			ExposedPorts: []string{"4222/tcp"},
			Cmd:          []string{"-js"},
			WaitingFor: wait.ForLog("Server is ready").
				WithStartupTimeout(30 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		return nil, fmt.Errorf("start nats container: %w", err)
	}
	address, err := containerAddress(ctx, container, "4222/tcp")
	if err != nil {
		_ = testcontainers.TerminateContainer(container)
		return nil, fmt.Errorf("resolve nats address: %w", err)
	}
	return &NATS{container: container, url: "nats://" + address}, nil
}

func (nats *NATS) URL() string {
	return nats.url
}

func (nats *NATS) Close(ctx context.Context) error {
	return testcontainers.TerminateContainer(nats.container)
}
