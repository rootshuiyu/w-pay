package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一 API 返回结构（交付文档：code=200 表示成功）
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

const (
	CodeSuccess      = 200
	CodeBadRequest   = 400
	CodeUnauthorized = 401
	CodeForbidden    = 403
	CodeNotFound     = 404
	CodeTooMany      = 429
	CodeServerError  = 500
)

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: CodeSuccess, Message: "success", Data: data})
}

func OKMessage(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{Code: CodeSuccess, Message: msg})
}

func Fail(c *gin.Context, httpCode, bizCode int, msg string) {
	c.JSON(httpCode, Response{Code: bizCode, Message: msg})
}

func BadRequest(c *gin.Context, msg string) {
	Fail(c, http.StatusBadRequest, CodeBadRequest, msg)
}

func Unauthorized(c *gin.Context, msg string) {
	Fail(c, http.StatusUnauthorized, CodeUnauthorized, msg)
}

func Forbidden(c *gin.Context, msg string) {
	Fail(c, http.StatusForbidden, CodeForbidden, msg)
}

func NotFound(c *gin.Context, msg string) {
	Fail(c, http.StatusNotFound, CodeNotFound, msg)
}

func TooManyRequests(c *gin.Context) {
	Fail(c, http.StatusTooManyRequests, CodeTooMany, "请求过于频繁，请稍后再试")
}

func ServerError(c *gin.Context, msg string) {
	Fail(c, http.StatusInternalServerError, CodeServerError, msg)
}
