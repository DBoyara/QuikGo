package quik

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Ping отправляет запрос ping на сервер.
func (q *QuikClient) Ping(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	request := getRequest()
	defer putRequest(request)

	request.Cmd = "ping"
	request.Data = "Ping"

	q.logger.Debug("sending ping request", zap.Any("request", request))

	response, err := q.client.sendRequest(ctx, request)
	if err != nil {
		return "", errors.Wrap(err, "failed to send ping request")
	}

	q.logger.Debug("received ping response", zap.String("message", response.Message))

	return response.Message, nil
}

// CreateDataSource создает источник данных для получения свечей.
func (q *QuikClient) CreateDataSource(data CreateDataSourceRequest, ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	request := getRequest()
	defer putRequest(request)

	request.Cmd = "createDataSource"
	request.Data = data

	q.logger.Debug("creating data source request", zap.String("ticker", data.Ticker), zap.Int("interval", data.Interval))

	response, err := q.client.sendRequest(ctx, request)
	if err != nil {
		return errors.Wrap(err, "failed to create data source")
	}

	if !response.Success {
		return fmt.Errorf("failed to create data source with message: %s", response.Message)
	}

	q.logger.Debug("data source created", zap.Bool("success", response.Success))

	return nil
}

// CreateDataSource закрывает источник данных.
func (q *QuikClient) CloseDataSource(data DataSourceRequest, ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	request := getRequest()
	defer putRequest(request)

	request.Cmd = "closeDataSource"
	request.Data = data

	q.logger.Debug("close data source request", zap.String("ticker", data.Ticker), zap.Int("interval", data.Interval))

	response, err := q.client.sendRequest(ctx, request)
	if err != nil {
		return errors.Wrap(err, "failed to close data source")
	}

	if !response.Success {
		return fmt.Errorf("failed to close data source with message: %s", response.Message)
	}

	q.logger.Debug("data source closed", zap.Bool("success", response.Success))

	return nil
}

func (q *QuikClient) SubscribeOrderBook(data SubscribeOrderBookRequest, ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	request := getRequest()
	defer putRequest(request)

	request.Cmd = "subscribeOrderBook"
	request.Data = data

	q.logger.Debug("subscribe OrderBook request", zap.String("classCode", data.ClassCode), zap.String("ticker", data.SecCode))

	response, err := q.client.sendRequest(ctx, request)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe OrderBook")
	}

	if !response.Success {
		return fmt.Errorf("failed to subscribe OrderBook with message: %s", response.Message)
	}

	q.logger.Debug("subscribe to OrderBook", zap.Bool("success", response.Success))

	return nil
}

// GetCandles возвращает свечи из источника данных.
func (q *QuikClient) GetCandles(data GetCandlesRequest, ctx context.Context) ([]Candle, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	request := getRequest()
	defer putRequest(request)

	request.Cmd = "getСandles"
	request.Data = data

	q.logger.Debug("getting candles request", zap.String("ticker", data.Ticker), zap.Int("interval", data.Interval))

	response, err := q.client.sendRequest(ctx, request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get candles: %w")
	}

	if !response.Success {
		return nil, fmt.Errorf("get candles failed with message: %s", response.Message)
	}

	q.logger.Debug("Received candles", zap.Any("response", response))

	var candles []Candle
	for _, item := range response.Candles {
		candles = append(candles, toCandleResult(item))
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

	response, err := q.client.sendRequest(ctx, request)
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

	response, err := q.client.sendRequest(ctx, request)
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
func (q *QuikClient) GetPortfolioInfo(data GetPortfolioRequest, ctx context.Context) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	request := getRequest()
	defer putRequest(request)

	request.Cmd = "getPortfolioInfo"
	request.Data = data

	q.logger.Debug("getting portfolio")

	response, err := q.client.sendRequest(ctx, request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get portfolio: %w")
	}

	if !response.Success {
		return nil, fmt.Errorf("get portfolio failed with message: %s", response.Message)
	}

	q.logger.Debug("Received portfolio", zap.Any("response", response))

	return response.Portfolio, nil
}

// SendTransaction - для работы с заявками.
// Варианты Action:
// NEW_ORDER - Новая лимитная/рыночная заявка
// Action KILL_ORDER - Удаление существующей заявки
// Action NEW_STOP_ORDER - Новая стоп заявка
// Action KILL_STOP_ORDER - Удаление существующей стоп-заявки
func (q *QuikClient) SendTransaction(data CreateOrderRequest, ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	request := getRequest()
	defer putRequest(request)

	request.Cmd = "sendTransaction"
	request.Data = data

	q.logger.Debug("send transaction", zap.Any("request", request))

	response, err := q.client.sendRequest(ctx, request)
	if err != nil {
		return errors.Wrap(err, "failed to send transaction: %w")
	}

	if !response.Success {
		return fmt.Errorf("send transaction failed with message: %s", response.Message)
	}

	q.logger.Debug("Received transaction response", zap.Any("response", response))

	return err
}

// GetOrderByNumber - возвращает заявку по режиму торгов и номеру
func (q *QuikClient) GetOrderByNumber(data GetOrderByNumberRequest, ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	request := getRequest()
	defer putRequest(request)

	request.Cmd = "getOrderByNumber"
	request.Data = data

	q.logger.Debug("getting order", zap.Any("request", request))

	response, err := q.client.sendRequest(ctx, request)
	if err != nil {
		return errors.Wrap(err, "failed to getting order: %w")
	}

	if !response.Success {
		return fmt.Errorf("getting order failed with message: %s", response.Message)
	}

	q.logger.Debug("Received order response", zap.Any("response", response))

	return err
}

// GetOrderById - возвращает заявку по тикеру и коду транзакции заявки
func (q *QuikClient) GetOrderById(data GetOrderByIdRequest, ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	request := getRequest()
	defer putRequest(request)

	request.Cmd = "getOrderById"
	request.Data = data

	q.logger.Debug("getting order", zap.Any("request", request))

	response, err := q.client.sendRequest(ctx, request)
	if err != nil {
		return errors.Wrap(err, "failed to getting order: %w")
	}

	if !response.Success {
		return fmt.Errorf("getting order failed with message: %s", response.Message)
	}

	q.logger.Debug("Received order response", zap.Any("response", response))

	return err
}

// GetStopOrders - возвращает список стоп-заявок по заданному инструменту
func (q *QuikClient) GetStopOrders(data GetStopOrderByTickerRequest, ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	request := getRequest()
	defer putRequest(request)

	request.Cmd = "getStopOrders"
	request.Data = data

	q.logger.Debug("getting order", zap.Any("request", request))

	response, err := q.client.sendRequest(ctx, request)
	if err != nil {
		return errors.Wrap(err, "failed to getting order: %w")
	}

	if !response.Success {
		return fmt.Errorf("getting order failed with message: %s", response.Message)
	}

	q.logger.Debug("Received order response", zap.Any("response", response))

	return err
}
