package reCache

import (
	"context"
	"errors"
	"fmt"
	"ginWeb/utils/database"
	"reflect"
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
	return resp.Val(), nil
}

func GetDel(namespace string, key string) (interface{}, error) {
	resp := database.Rdb.GetDel(context.Background(), formatter(namespace, key))
	if resp.Err() != nil {
		return nil, resp.Err()
	}
	return resp.Val(), nil
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

// CustomExecute TODO:反射实现自定义调用
func CustomExecute(method string, params ...interface{}) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {

		}
	}()
	obj := reflect.TypeOf(database.Rdb)
	for i := 0; i < obj.NumMethod(); i++ {
		f := obj.Method(i)
		if f.Name != method {
			continue
		}
		funcParams := make([]reflect.Value, len(params))
		for j := 0; i < len(params); i++ {
			funcParams[j] = reflect.ValueOf(params[j])
		}
		resp := f.Func.Call(funcParams)
		if len(resp) == 0 {
			return nil, fmt.Errorf("failed to call method %s", f.Name)
		}
		return resp, nil
	}
	return nil, fmt.Errorf("method '%s' of redis not found", method)
}
