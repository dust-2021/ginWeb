package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"ginWeb/config"
	"ginWeb/utils/exchange"
	"ginWeb/utils/loguru"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	Spot = "https://api3.binance.com"
	Swap = "https://fapi.binance.com"
)

// Binance 币安
type Binance struct {
	ApiKey     string
	Secret     string
	RecvWindow int
}

// Sign 生成签名
func (receiver *Binance) Sign(msg string) (string, error) {
	if receiver.Secret == "" || receiver.ApiKey == "" {
		err := errors.New("secret or binanceApi key is empty")
		return "", err
	}
	h := hmac.New(sha256.New, []byte(receiver.Secret))
	h.Write([]byte(msg))
	return hex.EncodeToString(h.Sum(nil)), nil
}

// 将请求参数格式化为url参数并设置签名
func (receiver *Binance) urlFormatter(data map[string]interface{}, sign bool) (string, error) {
	if data == nil {
		data = make(map[string]interface{})
	}
	result := ""
	// 添加当前毫秒时间戳
	if _, f := data["timestamp"]; !f && sign {
		data["timestamp"] = time.Now().UnixMilli()
	}
	// 添加延迟参数
	if sign {
		data["recvWindow"] = receiver.RecvWindow
	}
	for k, v := range data {
		item := fmt.Sprintf("%s=%v", k, v)
		if len(result) == 0 {
			result = item
			continue
		}
		result += "&" + item
	}
	if sign {
		signature, err := receiver.Sign(result)
		if err != nil {
			return "", err
		}
		result += "&signature=" + signature
	}
	if len(result) != 0 {
		result = "?" + result
	}
	return result, nil
}

// Request 同步请求
func (receiver *Binance) Request(item exchange.ExInterface) {
	// 签名请求参数
	urlParam, err := receiver.urlFormatter(item.ReqData(), item.Sign())
	if err != nil {
		item.SetResult(nil, err)
		return
	}

	totalUrl := item.Route() + urlParam
	req, err := http.NewRequest(item.Method(), totalUrl, nil)

	if err != nil {
		item.SetResult(nil, err)
		return
	}
	// binanceApi key 添加至请求头
	if item.Sign() {
		req.Header.Add("X-MBX-APIKEY", receiver.ApiKey)
	}
	timeout := receiver.RecvWindow
	if timeout == 0 {
		timeout = 5000
	}
	client := http.Client{Timeout: time.Duration(timeout) * time.Millisecond}
	// 使用代理
	if config.Conf.Exchange.Proxy != "" {
		trans, err := url.Parse(config.Conf.Exchange.Proxy)
		if err == nil {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(trans),
			}
		}
	}

	res, err := client.Do(req)

	if err != nil {
		item.SetResult(nil, err)
		loguru.Logu.Errorf("request failed %s", err.Error())
		return
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}
	if res.StatusCode != 200 {
		err = errors.New(fmt.Sprintf("http status code %d: %s", res.StatusCode, string(data)))
		item.SetResult(nil, err)
		loguru.Logu.Errorf("request binance %s failed: %s", totalUrl, err.Error())
		return
	}

	item.SetResult(data, nil)
	loguru.Logu.Infof("request binance %s success", totalUrl)
	return
}

func (receiver *Binance) AsyncRequests(reqs ...exchange.ExInterface) error {
	if len(reqs) == 0 {
		return errors.New("reqs is empty")
	}

	// 一次请求
	if len(reqs) == 1 {
		receiver.Request(reqs[0])
		return nil
	}
	wg := sync.WaitGroup{}

	for _, req := range reqs {
		wg.Add(1)
		req := req
		go func() {
			defer wg.Done()
			receiver.Request(req)
		}()
	}
	wg.Wait()

	return nil
}

func (receiver *Binance) SyncRequests(reqs ...exchange.ExInterface) error {
	if len(reqs) == 0 {
		return errors.New("reqs is empty")
	}
	for _, req := range reqs {
		receiver.Request(req)
	}
	return nil
}
