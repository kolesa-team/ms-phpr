package logger

import (
	"log"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/endeveit/go-snippets/config"
	"github.com/gemnasium/logrus-graylog-hook"
)

var (
	once   sync.Once
	logger *logrus.Logger
)

type nullFormatter struct {
}

func initLogger() {
	once.Do(func() {
		// Настройка логов
		logger = logrus.New()

		if graylogAddr, err := config.Instance().String("graylog", "address"); err == nil {
			extra := make(map[string]interface{})
			extra["hostname"], _ = os.Hostname()

			logger.Hooks.Add(graylog.NewGraylogHook(graylogAddr, extra))
		}

		log.SetOutput(logger.Writer())
	})
}

func Instance() *logrus.Logger {
	initLogger()

	return logger
}

// Не шлем логи в stdout
func (nullFormatter) Format(e *logrus.Entry) ([]byte, error) {
	return []byte{}, nil
}
