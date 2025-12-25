package loguru

import (
	"fmt"
	"ginWeb/config"
	"os"
	"strings"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

var Logger = &logrus.Logger{}
var DbLogger = Logger

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
	if config.Conf.Server.Debug {
		Logger.SetOutput(os.Stdout)
	} else {
		Logger.SetOutput(&lumberjack.Logger{
			Filename:   config.Conf.Server.Logger.Path + "/server.log",
			MaxSize:    100,
			MaxBackups: 10,
			MaxAge:     30,
			Compress:   true,
			LocalTime:  true,
		})
		// db日志分流
		DbLogger = &logrus.Logger{}
		DbLogger.SetOutput(&lumberjack.Logger{
			Filename:   config.Conf.Server.Logger.Path + "/db.log",
			MaxSize:    100,
			MaxBackups: 10,
			MaxAge:     30,
			Compress:   true,
			LocalTime:  true,
		})
		DbLogger.SetLevel(logrus.Level(level))
		DbLogger.SetFormatter(&myFormatter{})
	}
	Logger.SetLevel(logrus.Level(level))
	Logger.SetFormatter(&myFormatter{})
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
