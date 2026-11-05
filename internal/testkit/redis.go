//go:build integration

package testkit

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const redisImage = "redis:7.4-alpine"

type Redis struct {
	container testcontainers.Container
	url       string
}

func StartRedis(ctx context.Context) (*Redis, error) {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        redisImage,
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor: wait.ForLog("Ready to accept connections").
				WithStartupTimeout(30 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		return nil, fmt.Errorf("start redis container: %w", err)
	}
	address, err := containerAddress(ctx, container, "6379/tcp")
	if err != nil {
		_ = testcontainers.TerminateContainer(container)
		return nil, fmt.Errorf("resolve redis address: %w", err)
	}
	return &Redis{container: container, url: "redis://" + address + "/0"}, nil
}

func (redis *Redis) URL() string {
	return redis.url
}

func (redis *Redis) Close(ctx context.Context) error {
	return testcontainers.TerminateContainer(redis.container)
}
