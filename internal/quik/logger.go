package quikgo

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger создает новый логгер в зависимости от окружения.
func NewLogger(development bool) (*zap.Logger, error) {
	if development {
		// Логгер для разработки
		return zap.NewDevelopment()
	}

	// Логгер для продакшена
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // Форматирование времени в ISO8601
	return config.Build()
}
