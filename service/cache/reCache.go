package reCache

import (
	"context"
	"errors"
	"fmt"
	"ginWeb/utils/database"
	"time"
)

// Set TODO 缓存结果解析
func Set(namespace string, key string, value interface{}, ex uint) error {
	if ex == 0 {
		return errors.New("cache expired must be positive")
	}
	if namespace == "" {
		namespace = "default"
	}
	fmtKey := fmt.Sprintf("::%s::%s", namespace, key)
	resp := database.Rdb.SetEX(context.Background(), fmtKey, value, time.Duration(ex)*time.Second)
	if resp.Err() != nil {
		return resp.Err()
	}
	return nil
}

func Get(namespace string, key string) (interface{}, error) {
	if namespace == "" {
		namespace = "default"
	}
	fmtKey := fmt.Sprintf("::%s::%s", namespace, key)
	resp := database.Rdb.Get(context.Background(), fmtKey)
	if resp.Err() != nil {
		return nil, resp.Err()
	}
	return nil, nil
}

func GetDel(namespace string, key string) (interface{}, error) {
	if namespace == "" {
		namespace = "default"
	}
	fmtKey := fmt.Sprintf("::%s::%s", namespace, key)
	resp := database.Rdb.GetDel(context.Background(), fmtKey)
	if resp.Err() != nil {
		return nil, resp.Err()
	}
	return nil, nil
}

func Del(namespace string, key string) error {
	if namespace == "" {
		namespace = "default"
	}
	fmtKey := fmt.Sprintf("::%s::%s", namespace, key)
	resp := database.Rdb.Del(context.Background(), fmtKey)
	if resp.Err() != nil {
		return resp.Err()
	}
	return nil
}
