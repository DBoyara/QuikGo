package quik

import (
	"fmt"
)

// EventHandler — функция для обработки различных событий.
func EventHandler(event Event) {
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
