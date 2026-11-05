//go:build integration

package testkit

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const postgresImage = "postgres:18-alpine"

type Postgres struct {
	container testcontainers.Container
	dsn       string
}

func StartPostgres(ctx context.Context) (*Postgres, error) {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        postgresImage,
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "ledgergate",
				"POSTGRES_PASSWORD": "ledgergate",
				"POSTGRES_DB":       "ledgergate_test",
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		return nil, fmt.Errorf("start postgres container: %w", err)
	}
	address, err := containerAddress(ctx, container, "5432/tcp")
	if err != nil {
		_ = testcontainers.TerminateContainer(container)
		return nil, fmt.Errorf("resolve postgres address: %w", err)
	}
	dsn := (&url.URL{
		Scheme: "postgres",
		User:   url.UserPassword("ledgergate", "ledgergate"),
		Host:   address,
		Path:   "/ledgergate_test",
		RawQuery: url.Values{
			"sslmode": []string{"disable"},
		}.Encode(),
	}).String()
	return &Postgres{container: container, dsn: dsn}, nil
}

func (postgres *Postgres) DSN() string {
	return postgres.dsn
}

func (postgres *Postgres) Close(ctx context.Context) error {
	return testcontainers.TerminateContainer(postgres.container)
}

func containerAddress(ctx context.Context, container testcontainers.Container, port string) (string, error) {
	host, err := container.Host(ctx)
	if err != nil {
		return "", err
	}
	mapped, err := container.MappedPort(ctx, port)
	if err != nil {
		return "", err
	}
	return net.JoinHostPort(host, mapped.Port()), nil
}
