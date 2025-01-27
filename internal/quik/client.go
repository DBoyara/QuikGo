package quik

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Client — интерфейс для работы с TCP-клиентом.
type Client interface {
	SendRequest(ctx context.Context, request interface{}) (Response, error)
	Close() error
}

// QuikClient — клиент для работы с QUIK.
type QuikClient struct {
	client Client
	logger *zap.Logger
}

// NewQuikClient создает новый экземпляр QuikClient.
func NewQuikClient(host string, port int, isDevelopment bool) (*QuikClient, error) {
	tcpClient, err := NewTCPClient(host, port)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create TCP client")
	}

	logger, err := NewLogger(isDevelopment)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return &QuikClient{
		client: tcpClient,
		logger: logger,
	}, nil
}

// Ping отправляет запрос ping на сервер.
func (q *QuikClient) Ping(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	request := getRequest()
	defer putRequest(request)

	request.Cmd = "ping"
	request.Data = "Ping"

	q.logger.Debug("sending ping request", zap.Any("request", request))

	response, err := q.client.SendRequest(ctx, request)
	if err != nil {
		return "", errors.Wrap(err, "failed to send ping request")
	}

	q.logger.Debug("received ping response", zap.String("message", response.Message))

	return response.Message, nil
}

// CreateDataSource создает источник данных для получения свечей.
func (q *QuikClient) CreateDataSource(classCode, ticker string, interval int, ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	request := getRequest()
	defer putRequest(request)

	request.Cmd = "createDataSource"
	request.Data = CreateDataSourceRequest{
		Ticker:   ticker,
		Interval: interval,
		Class:    classCode,
	}

	q.logger.Debug("creating data source request", zap.String("classCode", classCode), zap.String("ticker", ticker), zap.Int("interval", interval))

	response, err := q.client.SendRequest(ctx, request)
	if err != nil {
		return errors.Wrap(err, "failed to create data source")
	}

	if !response.Success {
		return fmt.Errorf("failed to create data source with message: %s", response.Message)
	}

	q.logger.Debug("data source created", zap.Bool("success", response.Success))

	return nil
}

// GetCandles возвращает свечи из источника данных.
func (q *QuikClient) GetCandles(ticker, classCode string, interval, count int, ctx context.Context) ([]Candle, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	request := getRequest()
	defer putRequest(request)

	request.Cmd = "getСandles"
	request.Data = GetCandlesRequest{
		Class:    classCode,
		Ticker:   ticker,
		Interval: interval,
		Count:    count,
	}

	q.logger.Debug("getting candles", zap.String("ticker", ticker), zap.Int("count", count), zap.Int("interval", interval))

	response, err := q.client.SendRequest(ctx, request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get candles: %w")
	}

	if !response.Success {
		return nil, fmt.Errorf("get candles failed with message: %s", response.Message)
	}

	q.logger.Debug("Received candles", zap.Any("response", response))

	var candles []Candle
	for _, item := range response.Candles {
		candles = append(candles, ToCandleResult(item))
	}

	return candles, nil
}

// GetTradeAccounts
func (q *QuikClient) GetTradeAccounts(ctx context.Context) ([]Account, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	request := getRequest()
	defer putRequest(request)

	request.Cmd = "getTradeAccounts"

	q.logger.Debug("getting trade accounts")

	response, err := q.client.SendRequest(ctx, request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get trade accounts: %w")
	}

	if !response.Success {
		return nil, fmt.Errorf("get trade accounts failed with message: %s", response.Message)
	}

	q.logger.Debug("Received trade accounts", zap.Any("response", response))

	return response.Accounts, nil
}

// GetMoneyLimits
func (q *QuikClient) GetMoneyLimits(ctx context.Context) ([]MoneyLimits, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	request := getRequest()
	defer putRequest(request)

	request.Cmd = "getMoneyLimits"

	q.logger.Debug("getting money limits")

	response, err := q.client.SendRequest(ctx, request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get money limits: %w")
	}

	if !response.Success {
		return nil, fmt.Errorf("get money limits failed with message: %s", response.Message)
	}

	q.logger.Debug("Received money limits", zap.Any("response", response))

	return response.MoneyLimits, nil
}

// GetPortfolioInfo
func (q *QuikClient) GetPortfolioInfo(firmId, clientCode string, ctx context.Context) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	request := getRequest()
	defer putRequest(request)

	request.Cmd = "getPortfolioInfo"
	request.Data = GetPortfolioRequest{
		ClientCode: clientCode,
		FirmId:     firmId,
	}

	q.logger.Debug("getting portfolio")

	response, err := q.client.SendRequest(ctx, request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get portfolio: %w")
	}

	if !response.Success {
		return nil, fmt.Errorf("get portfolio failed with message: %s", response.Message)
	}

	q.logger.Debug("Received portfolio", zap.Any("response", response))

	return response.Portfolio, nil
}

// Close закрывает соединение с сервером.
func (q *QuikClient) Close() error {
	q.logger.Info("Closing Quik client")
	return q.client.Close()
}
