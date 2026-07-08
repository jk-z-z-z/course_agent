package errors

import "fmt"

type CodeError struct {
	Code    int
	Message string
}

func (e *CodeError) Error() string {
	return e.Message
}

func New(code int, message string) *CodeError {
	return &CodeError{Code: code, Message: message}
}

var (
	ErrInvalidParameter = New(40001, "参数校验失败")
	ErrUnauthorized     = New(41001, "未登录")
	ErrSessionExpired   = New(41002, "登录态已失效")
	ErrForbidden        = New(42001, "无权限")
	ErrUserExists       = New(40002, "用户名已存在")
	ErrUserNotFound     = New(43001, "用户不存在")
	ErrUserDisabled     = New(44001, "用户已被禁用")
)

func Wrap(code int, message string, err error) error {
	return fmt.Errorf("%w: %v", New(code, message), err)
}
