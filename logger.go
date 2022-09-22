package logNx

import (
	"encoding/json"
	"fmt"
	"github.com/Nixson/environment"
	"github.com/gol4ng/logger"
	"github.com/gol4ng/logger/formatter"
	"github.com/gol4ng/logger/handler"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

type customFormatter struct {
}

func (l *customFormatter) Format(entry logger.Entry) string {
	return fmt.Sprintf("%s %s %s", time.Now().Format(time.RFC1123Z), strings.ToUpper(entry.Level.String()), entry.Message)
}

var logLevel logger.Level

func SetLogLevel(loggingLevel string) {
	switch strings.ToLower(loggingLevel) {
	case "emergency":
		logLevel = logger.EmergencyLevel
	case "alert":
		logLevel = logger.AlertLevel
	case "critical":
		logLevel = logger.CriticalLevel
	case "error":
		logLevel = logger.ErrorLevel
	case "warning":
		logLevel = logger.WarningLevel
	case "notice":
		logLevel = logger.NoticeLevel
	case "info":
		logLevel = logger.InfoLevel
	case "debug":
		logLevel = logger.DebugLevel
	}
}
func IfEmergency() bool {
	return logLevel >= logger.EmergencyLevel
}

func IfAlert() bool {
	return logLevel >= logger.AlertLevel
}

func IfCritical() bool {
	return logLevel >= logger.CriticalLevel
}

func IfError() bool {
	return logLevel >= logger.ErrorLevel
}

func IfWarning() bool {
	return logLevel >= logger.WarningLevel
}

func IfNotice() bool {
	return logLevel >= logger.NoticeLevel
}

func IfInfo() bool {
	return logLevel >= logger.InfoLevel
}

func IfDebug() bool {
	return logLevel >= logger.DebugLevel
}

var lg *logger.Logger
var env *environment.Env

func LogInit(environment *environment.Env) {
	env = environment
	SetLogLevel(env.GetString("log.level"))
	var handlerStdOut = handler.Stream(os.Stdout, &customFormatter{})

	var handlerSocket = logger.NopHandler
	conn, err := net.DialTimeout("tcp",
		env.GetString("qsaver.host")+":"+env.GetString("qsaver.port"), 3*time.Second)
	if err != nil {
		log.Printf("Failed connect to %s:%s : '%s'\n",
			env.GetString("qsaver.host"),
			env.GetString("qsaver.port"),
			err.Error())
	} else {
		handlerSocket = handler.Socket(conn, formatter.NewJSON(func(entry logger.Entry) ([]byte, error) {
			host, err := os.Hostname()
			if err != nil {
				panic(err)
			}

			logMessage := FBLog{
				Time:         JSONTime(time.Now()),
				Level:        strings.ToUpper(entry.Level.String()),
				Message:      entry.Message,
				Error:        "_",
				Thread:       entry.Context.GoString(),
				Logger:       entry.Context.GoString(),
				Host:         host,
				Service:      env.GetString("service.name"),
				TraceId:      "_",
				SpanId:       "_",
				ParentSpanId: "_",
			}
			logMessageJson, err := json.Marshal(logMessage)
			if err != nil {
				return nil, err
			}
			return logMessageJson, nil
		}))

	}
	handlerGroup := handler.Group(
		logHandlerWithLevel(handlerStdOut),
		logHandlerWithLevel(handlerSocket),
	)

	lg = logger.NewLogger(handlerGroup)
}

func GetLogger() *logger.Logger {
	return lg
}
func logHandlerWithLevel(handlerInterface logger.HandlerInterface) logger.HandlerInterface {
	return func(entry logger.Entry) error {
		if logLevel >= entry.Level {
			err := handlerInterface(entry)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02T15:04:05.000Z"))
	return []byte(stamp), nil
}

type FBLog struct {
	Time         JSONTime `json:"time"`
	Level        string   `json:"level"`
	Message      string   `json:"message"`
	Error        string   `json:"error"`
	Thread       string   `json:"thread"`
	Logger       string   `json:"logger"`
	Host         string   `json:"host"`
	Service      string   `json:"service"`
	TraceId      string   `json:"traceId"`
	SpanId       string   `json:"spanId"`
	ParentSpanId string   `json:"parentSpanId"`
}
