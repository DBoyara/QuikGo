# QuikGo
Доступ к функционалу Quik из связки Lua&amp;Golang

![CI Status](https://github.com/DBoyara/QuikGo/actions/workflows/main.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/DBoyara/QuikGo)](https://goreportcard.com/report/github.com/DBoyara/QuikGo)
[![GoDoc](https://godoc.org/github.com/DBoyara/QuikGo?status.svg)](https://godoc.org/github.com/DBoyara/QuikGo)
[![License](https://img.shields.io/github/license/DBoyara/QuikGo)](https://github.com/DBoyara/QuikGo/blob/main/LICENSE)
[![GitHub release](https://img.shields.io/github/release/DBoyara/QuikGo.svg)](https://GitHub.com/DBoyara/QuikGo/releases/)
[![Known Vulnerabilities](https://snyk.io/test/github/DBoyara/QuikGo/badge.svg)](https://snyk.io/test/github/DBoyara/QuikGo)

## Install
go get -u github.com/DBoyara/QuikGo@v0.4.1

## Examples

### Пример подключения к серверу LUA и команды Ping
```go
func main() {
	logger, _ := zap.NewDevelopment()
	client, err := quik.NewQuikClient("127.0.0.1", 54320, true) // development = true — режим разработки, false — продакшен
	if err != nil {
		logger.Error("Failed to create Quik client", zap.Error(err))
		return
	}
	defer client.Close()

	context := context.Background()

	response, err := client.Ping(context)
	if err != nil {
		logger.Error("Failed to ping Quik server", zap.Error(err))
		return
	}

	logger.Info("Ping response", zap.Any("response", response))
}
```

### Пример создания DataSource и получения свечей
```go
func main() {
	dataDataSource := quik.CreateDataSourceRequest{
		Ticker:   "SBER",
		Interval: 1,
		Class:    "TQBR",
	}
	err := client.CreateDataSource(dataDataSource, context)
	if err != nil {
		logger.Error("Failed to create data source", zap.Error(err))
		return
	}

	dataCandles := quik.GetCandlesRequest{
		Ticker:   "SBER",
		Interval: 1,
		Count:    10,
		Class:    "TQBR",
	}
	candles, err := client.GetCandles(dataCandles, context)
	if err != nil {
		logger.Error("Failed to get candles", zap.Error(err))
		return
	}

	for _, candle := range candles {
		fmt.Printf("Candle: %s O: %.2f H: %.2f L: %.2f C: %.2f V: %d\n",
			candle.Timestamp,
			candle.Open,
			candle.High,
			candle.Low,
			candle.Close,
			candle.Volume,
		)
	}
}
```

### Пример получения аккаунтов
```go
accounts, err := client.GetTradeAccounts(context)
if err != nil {
    logger.Error("Failed to get trade accounts", zap.Error(err))
    return
}

fmt.Println(accounts)
```

### Пример денежных лимитов
```go
limits, err := client.GetMoneyLimits(context)
if err != nil {
    logger.Error("Failed to get limits", zap.Error(err))
    return
}

fmt.Println(limits)
```

### Account, ClientCode, FirmId - можно разово получить из GetTradeAccounts и GetMoneyLimits
### Пример заявок на покупку
```go
transID := quik.NewCounter(1)
dataOrder := quik.CreateOrderRequest{
    ClassCode: "TQBR",
    SecCode:   "SBER",
    Account:   "your-account-code",
    Trans_id:  transID.Next(),
    Operation: "B",
    Price:     "100.00",
    Quantity:  "1",
    Action:    "NEW_ORDER",
    Type:      "L",
}
err = client.SendTransaction(dataOrder, context)
if err != nil {
    logger.Error("Failed to create order", zap.Error(err))
    return
}
```

### Пример портфолио
```go
dataPortfolio := quik.GetPortfolioRequest{
    ClientCode: "your-client-code",
    FirmId:     "your-firm-id",
}
portfolio, err := client.GetPortfolioInfo(dataPortfolio, context)
if err != nil {
    logger.Error("Failed to get portfolio", zap.Error(err))
    return
}

fmt.Println(portfolio)
```

### Пример обработки Callback
```go
func eventHandler(event quik.Event) {
	switch event.Cmd {
	case "OnConnected":
		fmt.Println("✅ Подключение установлено:", event.Data)
	case "OnDisconnected":
		fmt.Println("❌ Подключение разорвано:", event.Data)
	case "OnTrade":
		fmt.Println("📊 Новая сделка:", event.Data)
	default:
		fmt.Println("⚠️ Неизвестное событие:", event.Cmd, "| Данные:", event.Data)
	}
}

func main() {
	logger, _ := quik.NewLogger(true)
	quik.RunServer(54321, true, eventHandler, logger)
}

```
