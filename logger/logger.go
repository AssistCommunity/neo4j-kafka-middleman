package logger

import (
	"os"

	"github.com/op/go-logging"
)

var log *logging.Logger

var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

func init() {
	log = logging.MustGetLogger("example")
	backend1 := logging.NewLogBackend(os.Stdout, "", 0)

	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	backend1Leveled := logging.AddModuleLevel(backend1Formatter)
	backend1Leveled.SetLevel(logging.INFO, "")

	logging.SetBackend(backend1Leveled)

}

func GetLogger() *logging.Logger {
	return log
}
