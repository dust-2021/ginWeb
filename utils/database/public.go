package database

import (
	"context"
	"fmt"
	"ginWeb/config"
	"ginWeb/utils/loguru"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"time"
)

type dbLogger struct {
	Logger *logrus.Logger
}

func (l *dbLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

func (l *dbLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	l.Logger.Infof("[GORM] | "+msg, args...)
}

func (l *dbLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	l.Logger.Warnf("[GORM] | "+msg, args...)
}

func (l *dbLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	l.Logger.Errorf("[GORM] | "+msg, args...)
}

func (l *dbLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if err != nil {
		l.Logger.Errorf("[GORM] | Error: %v", err)
	}
	sql, _ := fc()
	l.Logger.Infof("[GORM] | SQL: %s", sql)
}

// Db 数据库连接对象
var Db *gorm.DB

// Rdb redis连接对象
var Rdb *redis.Client

func init() {
	logu := &dbLogger{
		Logger: loguru.DbLogger,
	}
	db, err := gorm.Open(mysql.Open(config.Conf.Database.Link), &gorm.Config{
		Logger: logu,
		// 关闭复数形式表名
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		log.Panic("failed to connect database")
	}
	Db = db
	sqlDB, _ := Db.DB()
	sqlDB.SetMaxIdleConns(config.Conf.Database.PoolSize)
	sqlDB.SetMaxOpenConns(config.Conf.Database.MaxConnect)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		loguru.SimpleLog(loguru.Fatal, "DATABASE", "Failed to ping database: "+err.Error())
	} else {
		loguru.SimpleLog(loguru.Info, "DATABASE", "Successfully connected to the database")
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Conf.Redis.Host, config.Conf.Redis.Port),
		Password: config.Conf.Redis.Password,
		DB:       0,
		PoolSize: 10,
	})
	resp := rdb.Ping(context.Background())
	if err := resp.Err(); err != nil {
		loguru.SimpleLog(loguru.Fatal, "DATABASE", "Failed to ping redis: "+err.Error())
	} else {
		loguru.SimpleLog(loguru.Info, "DATABASE", "Successfully connected to the redis")
	}
	Rdb = rdb
}
