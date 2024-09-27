package wes

import (
	"encoding/json"
	"ginWeb/service/dataType"
	"ginWeb/utils/loguru"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
)

type handleFunc func(ctx *wContext)

var tasks map[string]handleFunc

type payload struct {
	Id        string   `json:"id"`
	Method    string   `json:"method"`
	Params    []string `json:"params"`
	Signature string   `json:"signature"`
}

type wContext struct {
	conn      *websocket.Conn
	data      payload
	attribute map[string]interface{}
}

func (w *wContext) Get(k string) (interface{}, bool) {
	attr, exist := w.attribute[k]
	return attr, exist
}

func (w *wContext) Handle() {
	if f, ok := tasks[w.data.Method]; ok {
		go f(w)
	} else {
		resp, _ := json.Marshal(w.data)
		err := w.conn.WriteMessage(websocket.BinaryMessage, resp)
		if err != nil {
			loguru.Logu.Errorf("return wes failed")
		}
	}

}

// 根据报文分配处理函数
func handlePayload(conn *websocket.Conn, message []byte) {
	var data payload
	err := json.Unmarshal(message, &data)
	if err != nil {
	}
	ctx := wContext{
		conn: conn,
		data: data,
	}
	ctx.Handle()
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
			break
		}
		if err != nil {
			resp := map[string]interface{}{}
			data, _ := json.Marshal(resp)
			_ = conn.WriteMessage(t, data)
			err := conn.Close()
			if err != nil {
				loguru.Logu.Errorf("close failed %s", err.Error())
			}
			break
		}
		go handlePayload(conn, message)

	}
}

func init() {
}
