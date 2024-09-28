package reCache

import (
	"context"
	"errors"
	"fmt"
	"ginWeb/utils/database"
	"time"
)

func formatter(namespace string, key string) string {
	if namespace == "" {
		namespace = "default"
	}
	return fmt.Sprintf("::%s::%s", namespace, key)
}

// Set 自动添加前缀
func Set(namespace string, key string, value interface{}, ex uint) error {
	if ex == 0 {
		return errors.New("cache expired must be positive")
	}
	resp := database.Rdb.SetEX(context.Background(), formatter(namespace, key), value, time.Duration(ex)*time.Second)
	if resp.Err() != nil {
		return resp.Err()
	}
	return nil
}

//func HSet(namespace string, setName string, value map[string]interface{}) error {
//	resp := database.Rdb.HSet(context.Background(), formatter(namespace, setName), value)
//	if resp.Err() != nil {
//		return resp.Err()
//	}
//	return nil
//}

func Get(namespace string, key string) (interface{}, error) {
	resp := database.Rdb.Get(context.Background(), formatter(namespace, key))
	if resp.Err() != nil {
		return nil, resp.Err()
	}
	return nil, nil
}

func GetDel(namespace string, key string) (interface{}, error) {
	resp := database.Rdb.GetDel(context.Background(), formatter(namespace, key))
	if resp.Err() != nil {
		return nil, resp.Err()
	}
	return nil, nil
}

func Del(namespace string, key string) error {
	resp := database.Rdb.Del(context.Background(), formatter(namespace, key))
	if resp.Err() != nil {
		return resp.Err()
	}
	return nil
}

func Incr(namespace string, key string) (int64, error) {
	resp := database.Rdb.Incr(context.Background(), formatter(namespace, key))
	if resp.Err() != nil {
		return 0, resp.Err()
	}
	return resp.Val(), nil
}

func Decr(namespace string, key string) (int64, error) {
	resp := database.Rdb.Decr(context.Background(), formatter(namespace, key))
	if resp.Err() != nil {
		return 0, resp.Err()
	}
	return resp.Val(), nil
}
