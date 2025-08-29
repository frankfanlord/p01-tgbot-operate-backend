package define

import "github.com/google/uuid"

const (
	CodeSuccess          uint32 = iota // 成功
	CodeUnAuthenticate                 // 未验证
	CodeParamError                     // 参数错误
	CodeSvrInternalError               // 服务内部错误
)

const (
	MsgUnknown          = "unknown error"
	MsgSuccess          = "success"
	MsgUnAuthenticate   = "unAuthenticate"
	MsgParamError       = "param error"
	MsgSvrInternalError = "server internal error"
)

// CodeText 状态码转错误信息
func CodeText(code uint32) string {
	switch code {
	case CodeSuccess:
		{
			return MsgSuccess
		}
	case CodeUnAuthenticate:
		{
			return MsgUnAuthenticate
		}
	case CodeParamError:
		{
			return MsgParamError
		}
	case CodeSvrInternalError:
		{
			return MsgSvrInternalError
		}
	}

	return MsgUnknown
}

// Response 响应
func Response(code uint32, data any) (string, BaseResponse) {
	return ResponseMsg(code, CodeText(code), data)
}

// ResponseMsg 响应带自定义msg
func ResponseMsg(code uint32, msg string, data any) (string, BaseResponse) {
	trace := ""
	if code != CodeSuccess {
		trace = uuid.New().String()
	}
	return trace, BaseResponse{
		Code:  code,
		Msg:   msg,
		Trace: trace,
		Data:  data,
	}
}

// BaseResponse 基础响应
type BaseResponse struct {
	Code  uint32 `json:"code"`
	Msg   string `json:"message"`
	Trace string `json:"trace,omitempty"`
	Data  any    `json:"data"`
}
