package auth

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"ginWeb/config"
	"io"
	"strings"
	"time"
)

// 字符串"{"alg": "hmac", "type": "jwt"}"的b64编码，作为JWT的header
const jwtHeader string = "eyJhbGciOiAiaG1hYyIsICJ0eXBlIjogImp3dCJ9"

// 去除扩展数据
func pkcs7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("数据长度错误")
	}
	padding := int(data[length-1])
	if padding > length || padding > aes.BlockSize {
		return nil, errors.New("无效扩展长度")
	}
	return data[:length-padding], nil
}

// 将数据扩展至固定长度
func pkcs7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	paddingText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, paddingText...)
}

type Token struct {
	UserId     int64                  `json:"userId"`
	UserUUID   string                 `json:"userUUID"`
	Permission []string               `json:"permission"`
	Data       map[string]interface{} `json:"data"`
	Expire     time.Time              `json:"expire"`
}

func (receiver *Token) hSign() (token string, err error) {
	data, err := json.Marshal(receiver)
	if err != nil {
		return
	}
	if len(data) > config.Conf.Server.TokenSize {
		err = errors.New("token data too long")
		return
	}

	h := hmac.New(sha256.New, []byte(config.Conf.Server.Secret))

	h.Write(data)

	// 计算 HMAC 值
	encryptedData := h.Sum(nil)

	// 将加密后的数据转换为b64字符串
	encryptedDataHex := base64.StdEncoding.EncodeToString(encryptedData)
	return fmt.Sprintf("%s.%s.%s", jwtHeader, base64.StdEncoding.EncodeToString(data), encryptedDataHex), nil
}

func (receiver *Token) Sign() (string, error) {
	if config.Conf.Server.TokenEncrypt {
		return receiver.AesEncrypt()

	}
	return receiver.hSign()
}

// checkSign 验证token并返回Token对象
func checkSign(t string) (token *Token, err error) {
	tokenChars := strings.Split(t, ".")
	if len(tokenChars) != 3 {
		err = errors.New("token format error")
		return
	}
	tokenBytes, err := base64.StdEncoding.DecodeString(tokenChars[1])
	if err != nil {
		return
	}
	err = json.Unmarshal(tokenBytes, &token)
	if err != nil {
		return
	}
	// 重新生成签名
	recheck, err := token.hSign()
	if err != nil {
		return
	}
	// 验证签名
	if recheck != t {
		err = errors.New("token sign error")
		return
	}

	return
}

// AesEncrypt 生成aes加密密文
func (receiver *Token) AesEncrypt() (token string, err error) {
	// 使用json数据加密
	data, err := json.Marshal(receiver)
	if len(data) > config.Conf.Server.TokenSize {
		err = errors.New("token data too long")
		return
	}
	data = pkcs7Padding(data, aes.BlockSize)
	if err != nil {
		return
	}
	t, err := aes.NewCipher([]byte(config.Conf.Server.Secret))
	if err != nil {
		return
	}
	ciphertext := make([]byte, aes.BlockSize+len(data))
	iv := ciphertext[:aes.BlockSize]
	// 生成 IV
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return
	}
	// 创建加密容器
	stream := cipher.NewCBCEncrypter(t, iv)

	stream.CryptBlocks(ciphertext[aes.BlockSize:], data)
	// 二进制转b64
	return base64.StdEncoding.EncodeToString(ciphertext), err
}

// 解密token数据
func aesDecrypt(tokenText string) (token *Token, err error) {
	data, err := base64.StdEncoding.DecodeString(tokenText)
	if err != nil {
		return
	}
	if len(data) < aes.BlockSize {
		err = errors.New("token长度错误")
		return
	}
	// 提取 IV
	iv := data[:aes.BlockSize]
	ciphertext := data[aes.BlockSize:]

	// 创建 AES CBC 解密器
	block, err := aes.NewCipher([]byte(config.Conf.Server.Secret))
	if err != nil {
		return
	}
	// 创建 CBC 解密器
	decrypt := cipher.NewCBCDecrypter(block, iv)
	// 解密数据
	decrypt.CryptBlocks(ciphertext, data[aes.BlockSize:])
	plaintext, err := pkcs7UnPadding(ciphertext)
	if err != nil {
		return
	}
	// 解析 JSON 数据
	err = json.Unmarshal(plaintext, &token)
	return
}

// CheckToken 检测token是否有效
func CheckToken(tokenText string) (token *Token, err error) {
	if config.Conf.Server.TokenEncrypt {
		token, err = aesDecrypt(tokenText)
	} else {
		token, err = checkSign(tokenText)
	}

	if err != nil {
		return
	}
	// 验证时间
	if time.Now().After(token.Expire) {
		return token, errors.New("token expired")
	}
	return
}
