package api

import (
	"github.com/gin-gonic/gin"
)

// ResponseCode 定义业务操作状态码。
// 区别于 HTTP 状态码，用于表示业务处理结果。
const (
	CodeSuccess        = 0    // 成功
	CodeInvalidParam   = 1001 // 参数校验失败
	CodeUnauthorized   = 1002 // 认证或授权失败
	CodeResourceExists = 1003 // 资源已存在（如用户名/地址已注册）
	CodeInternalError  = 9999 // 服务器内部错误
)

// Response 封装了 API 响应的统一结构体 (API Envelope)
type Response struct {
	// Code: 业务状态码 (非 HTTP 状态码)。0 表示成功。
	Code int `json:"code"`

	// Data: 响应的实际业务数据，可以是任意类型。
	Data interface{} `json:"data,omitempty"`

	// Message: 状态描述信息，便于调试和用户理解。
	Message string `json:"message"`
}

// Success 封装成功的响应，并设置 HTTP 状态码为 200/201 等
func Success(c *gin.Context, httpStatus int, data interface{}, message string) {
	if message == "" {
		message = "操作成功"
	}
	c.JSON(httpStatus, Response{
		Code:    CodeSuccess,
		Data:    data,
		Message: message,
	})
}

// Error 封装错误的响应，设置相应的 HTTP 状态码
func Error(c *gin.Context, httpStatus int, code int, message string) {
	if message == "" {
		message = "请求处理失败"
	}
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
	})
}
