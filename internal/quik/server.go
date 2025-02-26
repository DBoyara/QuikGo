package quik

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// QuikServer управляет TCP-сервером для получения callback из Lua
type QuikServer struct {
	listener     net.Listener
	logger       *zap.Logger
	mu           sync.Mutex
	clients      map[net.Conn]struct{}
	eventHandler func(event Event)
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewQuikServer создает новый экземпляр QuikServer
func NewQuikServer(port int, isDevelopment bool) (*QuikServer, error) {
	logger, err := NewLogger(isDevelopment)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create logger: %w")
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, errors.Wrap(err, "failed to start server: %w")
	}

	ctx, cancel := context.WithCancel(context.Background())

	server := &QuikServer{
		listener: listener,
		logger:   logger,
		clients:  make(map[net.Conn]struct{}),
		ctx:      ctx,
		cancel:   cancel,
	}

	logger.Debug("✅ QuikServer запущен", zap.Int("port", port))

	go server.acceptConnections()

	return server, nil
}

// acceptConnections слушает новые подключения от Lua
func (s *QuikServer) acceptConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				s.logger.Debug("🛑 Сервер закрыт, останавливаем acceptConnections")
				return
			default:
				s.logger.Error("❌ Ошибка при подключении клиента", zap.Error(err))
				continue
			}
		}

		s.logger.Debug("🔗 Подключен новый клиент", zap.String("remote", conn.RemoteAddr().String()))

		s.mu.Lock()
		s.clients[conn] = struct{}{}
		s.mu.Unlock()

		go s.handleClient(conn)
	}
}

// handleClient обрабатывает входящие сообщения от Lua
func (s *QuikServer) handleClient(conn net.Conn) {
	defer func() {
		s.logger.Debug("🔻 Отключен клиент", zap.String("remote", conn.RemoteAddr().String()))
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
		if err := conn.Close(); err != nil {
			s.logger.Error("❌ Ошибка отключения клиента", zap.String("remote", conn.RemoteAddr().String()), zap.Error(err))
		}
	}()

	reader := bufio.NewReader(conn)

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			msg, err := reader.ReadString('\n')
			if err != nil {
				if err.Error() == "EOF" {
					s.logger.Debug("🔻 Клиент закрыл соединение", zap.String("remote", conn.RemoteAddr().String()))
				} else {
					s.logger.Error("❌ Ошибка чтения от Lua", zap.Error(err))
				}
				return
			}

			var event Event
			if err := json.Unmarshal([]byte(msg), &event); err != nil {
				s.logger.Error("❌ Ошибка парсинга JSON", zap.String("raw", msg), zap.Error(err))
				continue
			}

			if s.eventHandler != nil {
				s.eventHandler(event)
			}
		}
	}
}

// SetEventHandler устанавливает кастомный обработчик событий
func (s *QuikServer) setEventHandler(handler func(event Event)) {
	s.eventHandler = handler
}

// Close завершает работу сервера и всех соединений
func (s *QuikServer) close() {
	s.logger.Debug("🛑 Закрываем QuikServer...")

	s.cancel()
	if err := s.listener.Close(); err != nil {
		s.logger.Error("❌ Ошибка закрытия QuikServer", zap.Error(err))
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for conn := range s.clients {
		if err := conn.Close(); err != nil {
			s.logger.Error("❌ Ошибка закрытия соединения", zap.String("remote", conn.RemoteAddr().String()), zap.Error(err))
		}
	}

	s.logger.Debug("✅ QuikServer полностью остановлен")
}

// RunServer запускает сервер и обрабатывает сигналы ОС (Ctrl+C, SIGTERM)
func RunServer(port int, isDevelopment bool, handler func(event Event), logger *zap.Logger) {
	server, err := NewQuikServer(port, isDevelopment)
	if err != nil {
		logger.Error("Ошибка запуска сервера:", zap.Error(err))
		return
	}

	server.setEventHandler(handler)

	// Обработка Ctrl+C и SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	logger.Debug("🛑 Получен сигнал завершения, закрываем сервер...")
	server.close()
}
