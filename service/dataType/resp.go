package dataType

type JsonRes struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

type JsonWrong struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
