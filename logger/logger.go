package logger

import (
	"os"
	"sync"

	"github.com/op/go-logging"
)

var doOnce sync.Once
var log *logging.Logger

var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

func initLogger(logLevel logging.Level) *logging.Logger {
	log := logging.MustGetLogger("logger")

	stdOutBackend := logging.NewLogBackend(os.Stdout, "", 0)

	stdOutBackendFormatter := logging.NewBackendFormatter(stdOutBackend, format)
	stdOutBackendLeveled := logging.AddModuleLevel(stdOutBackendFormatter)
	stdOutBackendLeveled.SetLevel(logLevel, "")

	return log
}

func GetLogger(logLevel logging.Level) *logging.Logger {
	doOnce.Do(func() {
		log = initLogger(logLevel)
	})

	return log
}
