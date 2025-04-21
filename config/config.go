package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type Config struct {
	Server struct {
		NodeId       uint8  `yaml:"nodeId"`       // 分布式节点ID
		Port         uint16 `yaml:"port"`         // 端口
		UdpPort      uint16 `yaml:"udpPort"`      // udp端口
		PprofPort    uint16 `yaml:"pprofPort"`    // pprof端口
		Secret       string `yaml:"secret"`       // 加密密钥
		Debug        bool   `yaml:"debug"`        //
		TokenEncrypt bool   `yaml:"tokenEncrypt"` // token是加密或签名
		TokenSize    int    `yaml:"tokenSize"`    // token最大长度
		TokenExpire  int    `yaml:"tokenExpire"`  // token过期时间
		Websocket    struct {
			WsLifeTime      uint32 `yaml:"wsLifeTime"`      // ws连接生命周期
			WsTaskTimeout   uint32 `yaml:"wsTaskTimeout"`   // ws处理超时
			WsLoginLifetime uint32 `yaml:"wsLoginLifetime"` // ws登录有效时间
			WsHeartbeat     uint32 `yaml:"wsHeartbeat"`     // ws心跳检测
			WsMaxWaiting    uint8  `yaml:"wsMaxWaiting"`    // ws单个连接最大等待处理数量
		} `yaml:"websocket"`
		Logger struct {
			Path  string `yaml:"path"`  // 日志文件位置
			Level string `yaml:"level"` // 日志等级
		}
		// 管理员账号信息
		AdminUser struct {
			Username string `yaml:"username"`
			Phone    string `yaml:"phone"`
			Email    string `yaml:"email"`
			Password string `yaml:"password"`
		} `yaml:"adminUser"`
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

	// 使用aes加密token时限制secret长度
	secretSize := len(Conf.Server.Secret)
	if Conf.Server.TokenEncrypt && !(secretSize == 32 || secretSize == 24 || secretSize == 16) {
		log.Fatal("secret key length must be between 32 and 24 and 16 while using token encrypt mode.")
	}
}
