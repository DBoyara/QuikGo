package quik

import (
	"github.com/DBoyara/QuikGo/internal/quik"
)

// Реэкспортируем типы
type QuikClient = quik.QuikClient
type CreateDataSourceRequest = quik.CreateDataSourceRequest
type GetPortfolioRequest = quik.GetPortfolioRequest
type CreateOrderRequest = quik.CreateOrderRequest
type GetOrderByNumberRequest = quik.GetOrderByNumberRequest
type GetOrderByIdRequest = quik.GetOrderByIdRequest
type GetStopOrderByTickerRequest = quik.GetStopOrderByTickerRequest
type MoneyLimits = quik.MoneyLimits
type Account = quik.Account
type GetCandlesRequest = quik.GetCandlesRequest
type QuikCandle = quik.QuikCandle
type QuikTime = quik.QuikTime
type Candle = quik.Candle

// Реэкспортируем функции
var NewQuikClient = quik.NewQuikClient
var NewLogger = quik.NewLogger
var NewCounter = quik.NewCounter
