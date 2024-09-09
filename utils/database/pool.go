package database

import (
	"context"
	"fmt"
	"ginWeb/config"
	"ginWeb/utils/loguru"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"time"
)

var Db *gorm.DB
var Rdb *redis.Client

func init() {
	db, err := gorm.Open(mysql.Open(config.Conf.Database.Link), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		log.Panic("failed to connect database")
	}
	Db = db
	sqlDB, _ := Db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		loguru.Logu.Info("Failed to ping database: %s", err.Error())
	} else {
		loguru.Logu.Info("Successfully connected to the database")
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Conf.Redis.Host, config.Conf.Redis.Port),
		Password: config.Conf.Redis.Password,
		DB:       0,
		PoolSize: 10,
	})
	resp := rdb.Ping(context.Background())
	if err := resp.Err(); err != nil {
		loguru.Logu.Error("Failed to ping redis: %s", err.Error())
	} else {
		loguru.Logu.Info("Successfully connected to redis")
	}
	Rdb = rdb
}
