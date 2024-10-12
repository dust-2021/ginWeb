package dataType

const (
	Success = 0
	// Unknown 未知错误
	Unknown = 1
	// WrongBody 错误报文
	WrongBody = 10001
	// Timeout 超时
	Timeout         = 10002
	NotFound        = 10003
	WsResolveFailed = 10004
	NoToken         = 10005
	WrongToken      = 10006
	BlackToken      = 10007
	// WrongData 错误的数据
	WrongData          = 10008
	AlreadyExist       = 10009
	IpLimited          = 10010
	RouteLimited       = 10011
	UserLimited        = 10012
	DeniedByPermission = 10013
)
