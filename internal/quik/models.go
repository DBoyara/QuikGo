package quik

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
type request struct {
	Cmd  string      `json:"cmd"`
	Data interface{} `json:"data"`
}

// CreateDataSourceRequest — данные для создания DataSource.
type CreateDataSourceRequest struct {
	Ticker   string `json:"ticker"`
	Interval int    `json:"interval"`
	Class    string `json:"class_code"`
}

// GetPortfolioRequest — данные для создания DataSource.
type GetPortfolioRequest struct {
	ClientCode string `json:"clientCode"`
	FirmId     string `json:"firmId"`
}

// CreateOrderRequest — данные для создания заявки.
type CreateOrderRequest struct {
	ClassCode  string `json:"CLASSCODE"`
	SecCode    string `json:"SECCODE"`
	Account    string `json:"ACCOUNT"`
	Trans_id   string `json:"TRANS_ID"`
	Operation  string `json:"OPERATION"`
	Price      string `json:"PRICE"`
	Quantity   string `json:"QUANTITY"`
	Action     string `json:"ACTION"`
	Type       string `json:"TYPE"` // L = лимитная заявка (по умолчанию), M = рыночная заявка
	StopPrice  string `json:"STOPPRICE,omitempty"`
	ExpiryDate string `json:"EXPIRY_DATE,omitempty"` // GTC - Срок действия до отмены
}

type GetOrderByNumberRequest struct {
	ClassCode string `json:"class_code"`
	OrderId   string `json:"order_id"`
}

type GetOrderByIdRequest struct {
	ClassCode string `json:"class_code"`
	SecCode   string `json:"sec_code"`
	TransId   string `json:"trans_id"`
}

type GetStopOrderByTickerRequest struct {
	ClassCode string `json:"class_code"`
	SecCode   string `json:"sec_code"`
}

// Response — структура для получения ответов от Lua-скрипта.
type response struct {
	Success     bool          `json:"success"`
	Message     string        `json:"message,omitempty"`
	Candles     []QuikCandle  `json:"candles,omitempty"`
	Accounts    []Account     `json:"accounts,omitempty"`
	MoneyLimits []MoneyLimits `json:"limits,omitempty"`
	Portfolio   interface{}   `json:"portfolio,omitempty"`
}

type MoneyLimits struct {
	ClientCode   string  `json:"client_code"`
	Currentbal   float64 `json:"currentbal"`
	Currentlimit float64 `json:"currentlimit"`
	Firmid       string  `json:"firmid"`
}

type Account struct {
	Firmid        string `json:"firmid"`
	Trdaccid      string `json:"trdaccid"`
	Main_Trdaccid string `json:"main_trdaccid"`
	Trdacc_type   int    `json:"trdacc_type"`
	Description   string `json:"description"`
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

// parseQuikTime преобразует QuikTime в строку формата timestamp с учетом московского времени.
func parseQuikTime(qt QuikTime) string {
	t := time.Date(
		qt.Year, time.Month(qt.Month), qt.Day,
		qt.Hour, qt.Min, qt.Sec, qt.Msec*int(time.Millisecond),
		MoscowLocation,
	)
	return t.Format(time.RFC3339)
}

// toCandleResult преобразует QuikCandle в Candle.
func toCandleResult(qc QuikCandle) Candle {
	return Candle{
		Timestamp: parseQuikTime(qc.Time),
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
		return &request{}
	},
}

// getRequest возвращает объект Request из пула.
func getRequest() *request {
	return requestPool.Get().(*request)
}

// putRequest возвращает объект Request в пул.
func putRequest(req *request) {
	requestPool.Put(req)
}

// Event — структура для обработки колбэков от QUIK.
type Event struct {
	Cmd  string      `json:"cmd"`
	T    int64       `json:"t"`
	Data interface{} `json:"data"`
}
