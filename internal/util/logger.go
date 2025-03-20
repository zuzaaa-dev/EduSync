package util

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func InitLogger(level string) {
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		Logger.Fatalf("Ошибка парсинга уровня логирования: %v", err)
	}

	Logger.SetLevel(logLevel)
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	Logger.SetOutput(os.Stdout)

	Logger.Info("Логирование инициализировано")
}
