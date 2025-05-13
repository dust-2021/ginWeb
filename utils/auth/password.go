package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"ginWeb/config"
)

// HashString sha256签名密码
func HashString(content string, solt ...string) string {
	var theSolt string = config.Conf.Server.Secret
	if len(solt) != 0 {
		theSolt = solt[0]
	}
	hash := sha256.New()
	hash.Write([]byte(theSolt + content))
	hashedPassword := hash.Sum(nil)
	return hex.EncodeToString(hashedPassword)
}
