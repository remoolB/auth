package consts

// 错误码定义
const (
	// 系统错误: 1000-1999
	ErrSystem       = 1000 // 系统错误
	ErrParams       = 1001 // 参数错误
	ErrUnauthorized = 1002 // 未授权
	ErrForbidden    = 1003 // 禁止访问
	ErrNotFound     = 1004 // 资源不存在

	// 用户相关错误: 2000-2999
	ErrUserNotExist       = 2000 // 用户不存在
	ErrUserAlreadyExist   = 2001 // 用户已存在
	ErrPasswordIncorrect  = 2002 // 密码错误
	ErrVerifyCodeExpired  = 2003 // 验证码已过期
	ErrVerifyCodeInvalid  = 2004 // 验证码无效
	ErrPermissionDenied   = 2005 // 权限不足
	ErrUserKicked         = 2006 // 用户已被踢出
	ErrCodeTooFrequent    = 2007 // 验证码发送过于频繁
	ErrAccountFrozen      = 2008 // 账号已被冻结
	ErrLoginLocked        = 2009 // 登录已被锁定
	ErrInvalidCredentials = 2010 // 账号或密码错误

	// 数据库错误: 3000-3999
	ErrDatabase = 3000 // 数据库错误
	ErrMongo    = 3001 // MongoDB错误
	ErrRedis    = 3002 // Redis错误

	// 鉴权错误: 4000-4999
	ErrTokenInvalid    = 4000 // Token无效
	ErrTokenExpired    = 4001 // Token已过期
	ErrTokenGenerating = 4002 // Token生成失败
	ErrTokenBlacklist  = 4003 // Token已被拉黑
)

// 错误信息映射
var ErrMsg = map[int]string{
	// 系统错误
	ErrSystem:       "系统错误",
	ErrParams:       "参数错误",
	ErrUnauthorized: "未授权",
	ErrForbidden:    "禁止访问",
	ErrNotFound:     "资源不存在",

	// 用户相关错误
	ErrUserNotExist:       "用户不存在",
	ErrUserAlreadyExist:   "用户已存在",
	ErrPasswordIncorrect:  "密码错误",
	ErrVerifyCodeExpired:  "验证码已过期",
	ErrVerifyCodeInvalid:  "验证码无效",
	ErrPermissionDenied:   "权限不足，需要管理员权限",
	ErrUserKicked:         "用户已被踢出",
	ErrCodeTooFrequent:    "验证码发送过于频繁，请稍后再试",
	ErrAccountFrozen:      "账号已被冻结，请30分钟后再试",
	ErrLoginLocked:        "登录失败次数过多，账号已被锁定，请30分钟后再试",
	ErrInvalidCredentials: "账号或密码错误",

	// 数据库错误
	ErrDatabase: "数据库错误",
	ErrMongo:    "MongoDB错误",
	ErrRedis:    "Redis错误",

	// 鉴权错误
	ErrTokenInvalid:    "无效的Token",
	ErrTokenExpired:    "Token已过期",
	ErrTokenGenerating: "Token生成失败",
	ErrTokenBlacklist:  "Token已被加入黑名单",
}

// ErrorWithCode 带错误码的错误接口
type ErrorWithCode interface {
	error
	ErrorCode() int
}

// AppError 应用错误结构体
type AppError struct {
	Code int    // 错误码
	Msg  string // 错误信息
}

// Error 实现error接口
func (e *AppError) Error() string {
	return e.Msg
}

// ErrorCode 获取错误码
func (e *AppError) ErrorCode() int {
	return e.Code
}

// NewAppError 创建应用错误
func NewAppError(code int, msg string) *AppError {
	return &AppError{
		Code: code,
		Msg:  msg,
	}
}

// NewAppErrorWithCode 根据错误码创建应用错误
func NewAppErrorWithCode(code int) *AppError {
	return &AppError{
		Code: code,
		Msg:  ErrMsg[code],
	}
}
