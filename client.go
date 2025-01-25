package quikgo

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

type Request struct {
	Cmd  string      `json:"cmd"`
	Data interface{} `json:"data"`
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

func main() {
	// Параметры соединения
	host := "127.0.0.1"
	port := "34130"
	conn, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		fmt.Printf("Ошибка подключения: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("Подключение установлено")

}

func sendRequest(conn net.Conn, request Request) (*Response, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации запроса: %w", err)
	}

	data = append(data, '\n')
	_, err = conn.Write(data)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %w", err)
	}

	reader := bufio.NewReader(conn)
	respData, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("ошибка получения ответа: %w", err)
	}

	var response Response
	err = json.Unmarshal([]byte(respData), &response)
	if err != nil {
		return nil, fmt.Errorf("ошибка десериализации ответа: %w", err)
	}

	return &response, nil
}
