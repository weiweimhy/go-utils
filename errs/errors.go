package errs

import "errors"

// 通用错误定义，可用于 errors.Is() 判断。
var (
	// ErrNotFound 资源不存在
	ErrNotFound = errors.New("resource not found")

	// ErrUnauthorized 未授权访问
	ErrUnauthorized = errors.New("unauthorized access")

	// ErrInvalidParam 参数无效
	ErrInvalidParam = errors.New("invalid parameter")

	// ErrConnection 连接失败
	ErrConnection = errors.New("connection failed")

	// ErrTimeout 操作超时
	ErrTimeout = errors.New("operation timed out")

	// ErrInternal 内部错误
	ErrInternal = errors.New("internal server error")

	// ErrNotInitialized 组件未初始化
	ErrNotInitialized = errors.New("component not initialized")

	// ErrClosed 组件已关闭
	ErrClosed = errors.New("component is closed")

	// ErrAlreadyExists 资源已存在
	ErrAlreadyExists = errors.New("resource already exists")

	// ErrEmpty 空数据
	ErrEmpty = errors.New("empty data")
)
