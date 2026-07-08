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
	ErrCourseExists     = New(40003, "课程编号已存在")
	ErrCourseMemberExists   = New(40004, "课程成员已存在")
	ErrMaterialExists       = New(40005, "资料节点已存在")
	ErrStorageQuotaExceeded = New(40006, "资料空间容量不足")
	ErrUserNotFound     = New(43001, "用户不存在")
	ErrCourseNotFound   = New(43002, "课程不存在")
	ErrCourseMemberNotFound = New(43003, "课程成员不存在")
	ErrMaterialNotFound     = New(43004, "资料不存在")
	ErrStorageSpaceNotFound = New(43005, "资料空间不存在")
	ErrAgentNotFound        = New(43006, "Agent 不存在")
	ErrConversationNotFound = New(43007, "会话不存在")
	ErrUserDisabled     = New(44001, "用户已被禁用")
	ErrAgentDisabled        = New(44002, "Agent 已禁用")
	ErrAgentUnavailable     = New(45001, "Agent 当前不可用")
)

func Wrap(code int, message string, err error) error {
	return fmt.Errorf("%w: %v", New(code, message), err)
}
