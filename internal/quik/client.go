package quik

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Client — интерфейс для работы с TCP-клиентом.
type Client interface {
	sendRequest(ctx context.Context, request interface{}) (response, error)
	close() error
}

// QuikClient — клиент для работы с QUIK.
type QuikClient struct {
	client Client
	logger *zap.Logger
	done   chan struct{}
}

// NewQuikClient создает новый экземпляр QuikClient.
func NewQuikClient(host string, port int, isDevelopment bool) (*QuikClient, error) {
	tcpClient, err := newTCPClient(host, port)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create TCP client")
	}

	logger, err := NewLogger(isDevelopment)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	client := &QuikClient{
		client: tcpClient,
		logger: logger,
		done:   make(chan struct{}),
	}

	return client, nil
}

// Close закрывает соединение с сервером.
func (q *QuikClient) Close() error {
	q.logger.Debug("🔻 Закрываем QUIK-клиент...")
	close(q.done)
	return q.client.close()
}
