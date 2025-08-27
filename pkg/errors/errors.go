package errors

import (
	"fmt"
	"runtime"

	"github.com/pkg/errors"
)

// ErrorCode 业务错误码
type ErrorCode int

const (
	// 通用错误码
	ErrCodeSuccess        ErrorCode = 0
	ErrCodeInternalError  ErrorCode = 500
	ErrCodeInvalidParam   ErrorCode = 400
	ErrCodeUnauthorized   ErrorCode = 401
	ErrCodeForbidden      ErrorCode = 403
	ErrCodeNotFound       ErrorCode = 404
	ErrCodeConflict       ErrorCode = 409
	ErrCodeTooManyRequest ErrorCode = 429

	// 业务错误码 (1000-9999)
	ErrCodeUserNotFound      ErrorCode = 1001
	ErrCodeUserAlreadyExists ErrorCode = 1002
	ErrCodeInvalidPassword   ErrorCode = 1003
	ErrCodeInvalidEmail      ErrorCode = 1004
)

// Error 业务错误结构
type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Err     error     `json:"-"`
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("code=%d, message=%s, error=%v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("code=%d, message=%s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// New 创建新的业务错误
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     errors.New(message),
	}
}

// Wrap 包装已有错误
func Wrap(err error, code ErrorCode, message string) *Error {
	if err == nil {
		return New(code, message)
	}
	return &Error{
		Code:    code,
		Message: message,
		Err:     errors.Wrap(err, message),
	}
}

// WithStack 添加调用栈信息
func WithStack(err error) error {
	if err == nil {
		return nil
	}
	return errors.WithStack(err)
}

// GetStackTrace 获取错误调用栈
func GetStackTrace(err error) []errors.Frame {
	if err == nil {
		return nil
	}
	var stackTracer interface {
		StackTrace() []errors.Frame
	}
	if errors.As(err, &stackTracer) {
		return stackTracer.StackTrace()
	}
	return nil
}

// Is 检查错误类型
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As 类型断言
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Cause 获取根本原因
func Cause(err error) error {
	return errors.Cause(err)
}

// 预定义错误
var (
	ErrInternalError  = New(ErrCodeInternalError, "internal server error")
	ErrInvalidParam   = New(ErrCodeInvalidParam, "invalid parameter")
	ErrUnauthorized   = New(ErrCodeUnauthorized, "unauthorized")
	ErrForbidden      = New(ErrCodeForbidden, "forbidden")
	ErrNotFound       = New(ErrCodeNotFound, "resource not found")
	ErrConflict       = New(ErrCodeConflict, "resource conflict")
	ErrTooManyRequest = New(ErrCodeTooManyRequest, "too many requests")

	ErrUserNotFound      = New(ErrCodeUserNotFound, "user not found")
	ErrUserAlreadyExists = New(ErrCodeUserAlreadyExists, "user already exists")
	ErrInvalidPassword   = New(ErrCodeInvalidPassword, "invalid password")
	ErrInvalidEmail      = New(ErrCodeInvalidEmail, "invalid email")
)

// 工具函数
func WrapInternalError(err error, message string) *Error {
	return Wrap(err, ErrCodeInternalError, message)
}

func WrapInvalidParam(err error, message string) *Error {
	return Wrap(err, ErrCodeInvalidParam, message)
}

func WrapNotFound(err error, message string) *Error {
	return Wrap(err, ErrCodeNotFound, message)
}

// 获取调用者信息
func GetCallerInfo(skip int) (string, int) {
	_, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "unknown", 0
	}
	return file, line
}
