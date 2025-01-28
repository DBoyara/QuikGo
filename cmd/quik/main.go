package main


import (
	"context"
	"fmt"

	"github.com/DBoyara/QuikGo/internal/quik"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()
	client, err := quik.NewQuikClient("127.0.0.1", 34130, true) // development = true — режим разработки, false — продакшен
	if err != nil {
		logger.Error("Failed to create Quik client", zap.Error(err))
		return
	}
	defer client.Close()

	context := context.Background()

	// Пример использования метода Ping
	response, err := client.Ping(context)
	if err != nil {
		logger.Error("Failed to ping Quik server", zap.Error(err))
		return
	}

	logger.Info("Ping response", zap.Any("response", response))
	// Пример создания источника данных
	err = client.CreateDataSource("TQBR", "SBER", 1, context)
	if err != nil {
		logger.Error("Failed to create data source", zap.Error(err))
		return
	}

	// Пример получения свечей
	candles, err := client.GetCandles("SBER", "TQBR", 1, 10, context)
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

	// Пример получения аккаунтов
	accounts, err := client.GetTradeAccounts(context)
	if err != nil {
		logger.Error("Failed to get trade accounts", zap.Error(err))
		return
	}

	fmt.Println(accounts)

	// Пример денежных лимитов
	limits, err := client.GetMoneyLimits(context)
	if err != nil {
		logger.Error("Failed to get limits", zap.Error(err))
		return
	}

	fmt.Println(limits)

	// Пример портфолио
	portfolio, err := client.GetPortfolioInfo("firmId", "clientCode", context)
	if err != nil {
		logger.Error("Failed to get portfolio", zap.Error(err))
		return
	}

	fmt.Println(portfolio)
}
