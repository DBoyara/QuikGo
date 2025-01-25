package quikgo

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/pkg/errors"
)

// TCPClient — реализация интерфейса Client для работы с TCP-соединением.
type TCPClient struct {
	conn   net.Conn
	reader *bufio.Reader
	mu     sync.Mutex
}

// NewTCPClient создает новый экземпляр TCPClient.
func NewTCPClient(host string, port int) (*TCPClient, error) {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	return &TCPClient{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}, nil
}

// SendRequest отправляет запрос на сервер и возвращает ответ.
func (c *TCPClient) SendRequest(ctx context.Context, request interface{}) (Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var response Response

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return response, errors.Wrap(err, "failed to marshal request")
	}

	_, err = c.conn.Write(append(requestBytes, '\n'))
	if err != nil {
		return response, errors.Wrap(err, "failed to send request")
	}

	responseBytes, err := c.reader.ReadBytes('\n')
	if err != nil {
		return response, errors.Wrap(err, "failed to read response")
	}

	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return response, errors.Wrap(err, "failed to unmarshal response")
	}

	return response, nil
}

// Close закрывает соединение с сервером.
func (c *TCPClient) Close() error {
	return c.conn.Close()
}
