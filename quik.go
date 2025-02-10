package quik

import (
	"github.com/DBoyara/QuikGo/internal/quik"
)

// Реэкспортируем типы
type QuikClient = quik.QuikClient
type QuikServer = quik.QuikServer
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
type Event = quik.Event
type Counter = quik.Counter

// Реэкспортируем функции
var NewQuikClient = quik.NewQuikClient
var NewQuikServer = quik.NewQuikServer
var NewLogger = quik.NewLogger
var NewCounter = quik.NewCounter
var RunServer = quik.RunServer
