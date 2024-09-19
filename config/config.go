package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type Config struct {
	Server struct {
		Port         uint16 `yaml:"port"`         // 端口
		Secret       string `yaml:"secret"`       // 加密密钥
		Debug        bool   `yaml:"debug"`        //
		TokenEncrypt bool   `yaml:"tokenEncrypt"` // token是加密或签名
		TokenSize    int    `yaml:"tokenSize"`    // token最大长度
		TokenExpire  int    `yaml:"tokenExpire"`  // token过期时间
		Logger       struct {
			Path  string `yaml:"path"`  // 日志文件位置
			Level string `yaml:"level"` // 日志等级
		}
	} `yaml:"server"`
	Database struct {
		Initial bool   `yaml:"initial"` // 是否生成表
		Link    string `yaml:"link"`    //
	} `yaml:"database"`
	Redis struct {
		Host     string `yaml:"host"`
		Port     uint16 `yaml:"port"`
		Password string `yaml:"password"`
	}
	// 交易所相关配置
	Exchange struct {
		Proxy string `yaml:"proxy"`
	}
}

var Conf *Config

func init() {
	data, err := os.ReadFile("./config.yaml")
	if err != nil {
		log.Fatalf("load config failed: %s", err.Error())
	}
	err = yaml.Unmarshal(data, &Conf)
	if err != nil {
		log.Fatalf("load config failed: %s", err.Error())
	}

}
