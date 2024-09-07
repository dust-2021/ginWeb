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
		TokenExpire  int    `yaml:"tokenExpire"`
	} `yaml:"server"`
	Database struct {
		Initial bool   `yaml:"initial"`
		Type    string `yaml:"type"`
		Link    string `yaml:"link"`
	} `yaml:"database"`
}

var Conf *Config

func init() {
	data, err := os.ReadFile("./config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(data, &Conf)
	if err != nil {
		log.Fatalf("load config failed: %s", err.Error())
	}

}
