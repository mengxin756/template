package response

import (
	"net/http"

	"example.com/classic/pkg/errors"
	"github.com/gin-gonic/gin"
)

// Response 统一响应格式
type Response struct {
	Code int         `json:"code"`           // 业务状态码
	Msg  string      `json:"msg"`            // 错误/成功信息
	Data interface{} `json:"data,omitempty"` // 返回数据
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: int(errors.ErrCodeSuccess),
		Msg:  "success",
		Data: data,
	})
}

// SuccessWithMsg 带消息的成功响应
func SuccessWithMsg(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: int(errors.ErrCodeSuccess),
		Msg:  msg,
		Data: data,
	})
}

// Error 错误响应
func Error(c *gin.Context, httpStatus int, err *errors.Error) {
	c.JSON(httpStatus, Response{
		Code: int(err.Code),
		Msg:  err.Message,
	})
}

// ErrorWithData 带数据的错误响应
func ErrorWithData(c *gin.Context, httpStatus int, err *errors.Error, data interface{}) {
	c.JSON(httpStatus, Response{
		Code: int(err.Code),
		Msg:  err.Message,
		Data: data,
	})
}

// 预定义响应函数
func BadRequest(c *gin.Context, err *errors.Error) {
	Error(c, http.StatusBadRequest, err)
}

func Unauthorized(c *gin.Context, err *errors.Error) {
	Error(c, http.StatusUnauthorized, err)
}

func Forbidden(c *gin.Context, err *errors.Error) {
	Error(c, http.StatusForbidden, err)
}

func NotFound(c *gin.Context, err *errors.Error) {
	Error(c, http.StatusNotFound, err)
}

func Conflict(c *gin.Context, err *errors.Error) {
	Error(c, http.StatusConflict, err)
}

func TooManyRequests(c *gin.Context, err *errors.Error) {
	Error(c, http.StatusTooManyRequests, err)
}

func InternalServerError(c *gin.Context, err *errors.Error) {
	Error(c, http.StatusInternalServerError, err)
}

// 便捷响应函数
func InvalidParam(c *gin.Context, message string) {
	BadRequest(c, errors.New(errors.ErrCodeInvalidParam, message))
}

func UserNotFound(c *gin.Context) {
	NotFound(c, errors.ErrUserNotFound)
}

func UserAlreadyExists(c *gin.Context) {
	Conflict(c, errors.ErrUserAlreadyExists)
}

func InvalidPassword(c *gin.Context) {
	BadRequest(c, errors.ErrInvalidPassword)
}

func InvalidEmail(c *gin.Context) {
	BadRequest(c, errors.ErrInvalidEmail)
}

// 分页响应
type PageResponse struct {
	Total      int64       `json:"total"`       // 总记录数
	Page       int         `json:"page"`        // 当前页码
	PageSize   int         `json:"page_size"`   // 每页大小
	TotalPages int         `json:"total_pages"` // 总页数
	HasNext    bool        `json:"has_next"`    // 是否有下一页
	HasPrev    bool        `json:"has_prev"`    // 是否有上一页
	Data       interface{} `json:"data"`        // 数据列表
}

// SuccessWithPage 分页成功响应
func SuccessWithPage(c *gin.Context, data interface{}, total int64, page, pageSize int) {
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	pageResp := PageResponse{
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
		Data:       data,
	}
	Success(c, pageResp)
}
