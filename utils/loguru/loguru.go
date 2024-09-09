package loguru

import (
	"fmt"
	"ginWeb/config"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

var Logu *logrus.Logger

type MyFormatter struct {
}

func (f *MyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	logMessage := fmt.Sprintf("[%s] | %s\n",
		strings.ToUpper(entry.Level.String()),
		entry.Message)
	return []byte(logMessage), nil
}

func init() {
	levelMap := map[string]uint32{
		"panic": 0, "fatal": 1, "error": 2, "warn": 3, "info": 4, "debug": 5, "trace": 6,
	}
	var level uint32
	level, f := levelMap[config.Conf.Server.Logger.Level]
	if !f {
		level = 4
	}
	var file *os.File
	if config.Conf.Server.Debug {
		file = os.Stdout
	} else {
		f, err := os.OpenFile("./loguru.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		file = f
		if err != nil {
			panic(err)
		}
	}
	Logu = &logrus.Logger{
		Out:       file,
		Formatter: &MyFormatter{},
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.Level(level),
	}
	Logu.Info("logrus configurate as %s", logrus.Level(level))
}
