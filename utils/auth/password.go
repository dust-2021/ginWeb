package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"ginWeb/config"
)

// HashPassword sha256签名密码
func HashPassword(password string) (string, error) {
	hash := sha256.New()
	hash.Write([]byte(config.Conf.Server.Secret + password))
	hashedPassword := hash.Sum(nil)
	return hex.EncodeToString(hashedPassword), nil
}
