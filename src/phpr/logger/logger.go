package logger

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/endeveit/go-snippets/config"
	"gopkg.in/gemnasium/logrus-graylog-hook.v1"
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

		if graylogHost, err := config.Instance().String("graylog", "host"); err == nil {
			if graylogPort, err := config.Instance().Int("graylog", "port"); err == nil {
				extra := make(map[string]interface{})
				extra["hostname"], _ = os.Hostname()

				logger.Hooks.Add(graylog.NewGraylogHook(fmt.Sprintf("%s:%d", graylogHost, graylogPort), "ms-phpr", extra))
			}
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
