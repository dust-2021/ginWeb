package reCache

import (
	"context"
	"errors"
	"fmt"
	"ginWeb/utils/database"
	"ginWeb/utils/loguru"
	"time"

	"github.com/go-redis/redis/v8"
)

var defaultExpire uint = 5 * 60
var defaultLockExpire uint = 2

func formatter(namespace string, key string) string {
	if namespace == "" {
		namespace = "default"
	}
	return fmt.Sprintf("::%s::%s", namespace, key)
}

// Set 自动添加前缀
func Set(namespace string, key string, value any, ex uint) error {
	if ex == 0 {
		ex = defaultExpire
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	resp := database.Rdb.SetEX(ctx, formatter(namespace, key), value, time.Duration(ex)*time.Second)
	if resp.Err() != nil {
		return resp.Err()
	}
	return nil
}

func regetCache(key string, f func() (any, error)) (any, error) {
	ctx, c := context.WithTimeout(context.Background(), 100*time.Microsecond)
	defer c()
	resp := database.Rdb.SetNX(ctx, key+"::_lock", true, time.Duration(defaultLockExpire)*time.Second)
	if resp.Err() != nil {
		return nil, resp.Err()
	}
	// 拿到锁
	if resp.Val() {
		defer database.Rdb.Del(context.Background(), key+"::_lock")
		newData, err := f()
		if err != nil {
			return nil, err
		}
		resetCtx, rec := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer rec()
		resp := database.Rdb.Set(resetCtx, key, newData, time.Duration(defaultExpire)*time.Second)
		if resp.Err() == nil {
			loguru.SimpleLog(loguru.Info, "CACHE", fmt.Sprintf("update cache data %s", key))
		}
		return newData, nil
	}
	return nil, errors.New("another worker is writing")
}

// 获取缓存数据，不存在则会重新加载
func Get(namespace string, key string, reget func() (any, error), ex uint) (any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	name := formatter(namespace, key)
	resp := database.Rdb.Get(ctx, name)
	if resp.Err() == redis.Nil && reget != nil {
		return regetCache(name, reget)
	}
	if resp.Err() != nil {
		return nil, resp.Err()
	}
	return resp.Val(), nil
}

func Incr(namespace string, key string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	resp := database.Rdb.Incr(ctx, formatter(namespace, key))
	if resp.Err() != nil {
		return 0, resp.Err()
	}
	return resp.Val(), nil
}

func Decr(namespace string, key string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	resp := database.Rdb.Decr(ctx, formatter(namespace, key))
	if resp.Err() != nil {
		return 0, resp.Err()
	}
	return resp.Val(), nil
}
