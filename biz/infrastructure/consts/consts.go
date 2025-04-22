package consts

const (
	// 通用常量
	Success = 0
	Failed  = -1

	// 验证码相关
	CodeLength      = 6            // 验证码长度
	CodeExpire      = 60 * 5       // 验证码过期时间，5分钟
	CodeRedisPrefix = "auth:code:" // 验证码Redis前缀

	// 验证码发送频率限制
	CodeCooldownPrefix  = "auth:cooldown:"   // 验证码冷却前缀
	CodeFirstCooldown   = 30                 // 首次发送后冷却时间，30秒
	CodeSecondCooldown  = 60                 // 第二次发送后冷却时间，60秒
	CodeFailCountPrefix = "auth:fail_count:" // 验证码失败次数前缀
	CodeMaxFailCount    = 5                  // 最大失败次数
	CodeFreezePrefix    = "auth:freeze:"     // 账号冻结前缀
	CodeFreezeTime      = 60 * 30            // 账号冻结时间，30分钟

	// 登录失败限制
	LoginFailEmailPrefix = "auth:login_fail:email:" // 登录失败邮箱前缀
	LoginFailIPPrefix    = "auth:login_fail:ip:"    // 登录失败IP前缀
	LoginLockEmailPrefix = "auth:login_lock:email:" // 登录锁定邮箱前缀
	LoginLockIPPrefix    = "auth:login_lock:ip:"    // 登录锁定IP前缀
	LoginMaxFailCount    = 5                        // 最大登录失败次数
	LoginFailExpire      = 60 * 60 * 24             // 登录失败记录过期时间，24小时
	LoginLockTime        = 60 * 30                  // 登录锁定时间，30分钟

	// 用户相关
	UserCollection       = "users"       // 用户集合名
	CredentialCollection = "credentials" // 登录凭证集合名

	// 角色相关
	RoleAdmin = "admin" // 管理员角色
	RoleUser  = "user"  // 普通用户角色

	// JWT相关
	TokenType   = "Bearer"        // Token类型
	TokenHeader = "Authorization" // 请求头名称

	// Token黑名单相关
	TokenBlacklistPrefix = "auth:blacklist:" // Token黑名单前缀
	TokenBlacklistExpire = 60 * 60 * 24 * 7  // Token黑名单过期时间，7天

	// MongoDB相关
	MongoTimeout = 10 // MongoDB操作超时时间(秒)
)
