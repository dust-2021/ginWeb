package loguru

import (
	"fmt"
	"ginWeb/config"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger
var DbLogger *logrus.Logger

const (
	Panic = logrus.PanicLevel
	Fatal = logrus.FatalLevel
	Error = logrus.ErrorLevel
	Warn  = logrus.WarnLevel
	Info  = logrus.InfoLevel
	Debug = logrus.DebugLevel
	Trace = logrus.TraceLevel
)

type myFormatter struct {
}

func (f *myFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	logMessage := fmt.Sprintf("[%-7s] | [%s] | %s\n",
		strings.ToUpper(entry.Level.String()),
		entry.Time.Format("2006-01-02 15:04:05.0000"),
		entry.Message)
	return []byte(logMessage), nil
}

func init() {
	levelMap := map[string]uint32{
		"panic": 0, "fatal": 1, "error": 2, "warn": 3, "info": 4, "debug": 5, "trace": 6,
	}
	var level uint32
	level, f := levelMap[strings.ToLower(config.Conf.Server.Logger.Level)]
	if !f {
		level = 4
	}
	var file *os.File
	if config.Conf.Server.Debug {
		file = os.Stdout
	} else {
		f, err := os.OpenFile(config.Conf.Server.Logger.Path+"/server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		file = f
		if err != nil {
			panic(err)
		}
	}
	Logger = &logrus.Logger{
		Out:       file,
		Formatter: &myFormatter{},
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.Level(level),
	}
	// 非调试下gorm日志放入不同文件
	if config.Conf.Server.Debug {
		DbLogger = Logger
	} else {
		f, err := os.OpenFile(config.Conf.Server.Logger.Path+"/db.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		DbLogger = &logrus.Logger{
			Out:       f,
			Formatter: &myFormatter{},
			Hooks:     make(logrus.LevelHooks),
			Level:     logrus.Level(level),
		}
	}
	SimpleLog(Info, "SYSTEM", fmt.Sprintf("logrus configurate as %s", logrus.Level(level)))
}

// SimpleLog 简易分类日志
func SimpleLog(level logrus.Level, type_ string, message string) {
	switch level {
	case logrus.PanicLevel:
		Logger.Panic(message)
	case logrus.FatalLevel:
		Logger.Fatal(message)
	default:
		Logger.Log(level, fmt.Sprintf("[%s] | %s", type_, message))
	}

}
