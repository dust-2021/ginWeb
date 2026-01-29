package dataType

const (
	Success = 0
	// Unknown 未知错误
	Unknown = 1
	// Forbidden 禁止访问
	Forbidden = 2
	// WrongBody 错误报文
	WrongBody = 10001
	// WrongData 数据内容错误
	WrongData = 10002
	// Timeout 超时
	Timeout         = 10003
	NotFound        = 10004
	TooManyRequests = 10005
	AlreadyExist    = 10006

	NoToken            = 10101
	WrongToken         = 10102
	BlackToken         = 10103
	DeniedByPermission = 10104
	IpLimited          = 10105
	RouteLimited       = 10106
	UserLimited        = 10107

	WsResolveFailed = 10201
	WsDuplicateAuth = 10202
)
