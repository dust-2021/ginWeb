package wes

import (
	"encoding/json"
	"ginWeb/service/dataType"
	"ginWeb/utils/loguru"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
)

type payload struct {
	Id        string   `json:"id"`
	Method    string   `json:"method"`
	Params    []string `json:"params"`
	Signature string   `json:"signature"`
}

type WContext struct {
	Conn *websocket.Conn
}

type handleFunc func(ctx *WContext)

type Ex interface {
	Handle(*WContext)
}

// 根据报文分配处理函数
func handlePayload(message []byte) error {
	var data payload
	err := json.Unmarshal(message, &data)
	if err != nil {
		return err
	}
	return nil
}

var upper = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		loguru.Logu.Errorf("wes from %s error: %s", r.RemoteAddr, reason.Error())
	},
}

func UpgradeConn(c *gin.Context) {
	conn, err := upper.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code: 1, Message: "upgrade failed",
		})
		return
	}
	for {
		t, message, err := conn.ReadMessage()
		if t == websocket.CloseMessage {
			loguru.Logu.Infof("close from %s", conn.RemoteAddr())
		}
		if err != nil {

		}
		err = handlePayload(message)
	}
}

func init() {
}
