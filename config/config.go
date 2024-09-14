package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type Config struct {
	Server struct {
		Port         uint16 `yaml:"port"`
		Secret       string `yaml:"secret"`
		Debug        bool   `yaml:"debug"`
		TokenEncrypt bool   `yaml:"tokenEncrypt"`
		TokenSize    int    `yaml:"tokenSize"`
		TokenExpire  int    `yaml:"tokenExpire"`
		Logger       struct {
			Path  string `yaml:"path"`
			Level string `yaml:"level"`
		}
	} `yaml:"server"`
	Database struct {
		Initial bool   `yaml:"initial"`
		Link    string `yaml:"link"`
	} `yaml:"database"`
	Redis struct {
		Host     string `yaml:"host"`
		Port     uint16 `yaml:"port"`
		Password string `yaml:"password"`
	}
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
