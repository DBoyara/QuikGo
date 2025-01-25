package quikgo

import (
	"fmt"
	"sync"
	"time"
)

var (
	// MoscowLocation — объект для работы с московским временем.
	MoscowLocation *time.Location
	once           sync.Once
)

func init() {
	once.Do(func() {
		var err error
		MoscowLocation, err = time.LoadLocation("Europe/Moscow")
		if err != nil {
			panic(fmt.Sprintf("failed to load Moscow time location: %v", err))
		}
	})
}

// Request — структура для отправки запросов на Lua-скрипт.
type Request struct {
	Cmd  string      `json:"cmd"`
	Data interface{} `json:"data"`
}

// CreateDataSourceRequest — данные для создания DataSource.
type CreateDataSourceRequest struct {
	Ticker   string `json:"ticker"`
	Interval int    `json:"interval"`
	Class    string `json:"class_code"`
}

// Response — структура для получения ответов от Lua-скрипта.
type Response struct {
	Success bool         `json:"success"`
	Message string       `json:"message,omitempty"`
	Candles []QuikCandle `json:"candles,omitempty"`
}

// GetCandlesRequest — данные для получения свечей.
type GetCandlesRequest struct {
	Class    string `json:"class_code"`
	Ticker   string `json:"ticker"`
	Interval int    `json:"interval"`
	Count    int    `json:"count"`
}

// QuikCandle — структура для представления свечи из QUIK.
type QuikCandle struct {
	Time   QuikTime `json:"time"`
	Open   float64  `json:"open"`
	Close  float64  `json:"close"`
	High   float64  `json:"high"`
	Low    float64  `json:"low"`
	Volume float64  `json:"volume"`
}

// QuikTime — структура для представления времени из QUIK.
type QuikTime struct {
	Count   int `json:"count"`
	Day     int `json:"day"`
	Hour    int `json:"hour"`
	Min     int `json:"min"`
	Month   int `json:"month"`
	Msec    int `json:"ms"`
	Sec     int `json:"sec"`
	WeekDay int `json:"week_day"`
	Year    int `json:"year"`
}

// Candle — структура для представления свечи в удобном формате.
type Candle struct {
	Timestamp string  `json:"timestamp"`
	Open      float64 `json:"open"`
	Close     float64 `json:"close"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Volume    int     `json:"volume"`
}

// ParseQuikTime преобразует QuikTime в строку формата timestamp с учетом московского времени.
func ParseQuikTime(qt QuikTime) string {
	t := time.Date(
		qt.Year, time.Month(qt.Month), qt.Day,
		qt.Hour, qt.Min, qt.Sec, qt.Msec*int(time.Millisecond),
		MoscowLocation,
	)
	return t.Format(time.RFC3339)
}

// ToCandleResult преобразует QuikCandle в Candle.
func ToCandleResult(qc QuikCandle) Candle {
	return Candle{
		Timestamp: ParseQuikTime(qc.Time),
		Open:      qc.Open,
		Close:     qc.Close,
		High:      qc.High,
		Low:       qc.Low,
		Volume:    int(qc.Volume),
	}
}

// sync.Pool для повторного использования объектов Request.
var requestPool = sync.Pool{
	New: func() interface{} {
		return &Request{}
	},
}

// getRequest возвращает объект Request из пула.
func getRequest() *Request {
	return requestPool.Get().(*Request)
}

// putRequest возвращает объект Request в пул.
func putRequest(req *Request) {
	requestPool.Put(req)
}
