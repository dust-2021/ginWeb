package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"math/rand"
	"strings"
)

func ContainsAll(a, b []string) bool {
	if len(b) == 0 {
		return true
	}
	if len(a) == 0 {
		return false
	}
	elementMap := make(map[string]struct{})
	for _, v := range a {
		elementMap[v] = struct{}{}
	}

	for _, v := range b {
		if _, found := elementMap[v]; !found {
			return false
		}
	}
	return true
}

func ContainOne(a []string, b [][]string) bool {
	if len(b) == 0 {
		return true
	}
	if len(a) == 0 {
		return false
	}
	elementMap := make(map[string]struct{})
	for _, v := range a {
		elementMap[v] = struct{}{}
	}
	for _, v := range b {
		flag := false
		for _, v1 := range v {
			if _, found := elementMap[v1]; found {
				flag = true
				break
			}
		}
		// 某个多选一未通过
		if !flag {
			return false
		}
	}
	return true
}

var jsonFormatter *validator.Validate

// ShouldBindJson json解析自定义数据类型标签
func ShouldBindJson(data []byte, target interface{}) error {
	err := json.Unmarshal(data, target)
	if err != nil {
		return err
	}
	err = jsonFormatter.Struct(target)
	if err != nil {
		var validationErrors validator.ValidationErrors
		switch {
		case errors.As(err, &validationErrors):
			var infos []string
			for _, e := range validationErrors {
				infos = append(infos, fmt.Sprintf("field %s failed on %s tag", e.Field(), e.Tag()))
			}
			return errors.New(strings.Join(infos, ","))
		default:
			return err
		}
	}
	return nil
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateRandomString 生成指定长度的随机字符串
func GenerateRandomString(length int) string {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = charset[rand.Intn(len(charset))]
	}

	return string(result)
}

func init() {
	jsonFormatter = validator.New()
}
