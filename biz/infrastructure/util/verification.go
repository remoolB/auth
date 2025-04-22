package util

import (
	"auth/biz/infrastructure/consts"
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// 初始化随机数生成器种子
func init() {
	rand.Seed(time.Now().UnixNano())
}

// GenerateVerificationCode 生成验证码
func GenerateVerificationCode() string {
	// 生成6位随机数字
	code := rand.Intn(900000) + 100000
	return strconv.Itoa(code)
}

// GetCodeRedisKey 获取验证码在Redis中的键
func GetCodeRedisKey(identifier string) string {
	return consts.CodeRedisPrefix + identifier
}

// GetCodeCooldownKey 获取验证码冷却在Redis中的键
func GetCodeCooldownKey(identifier string) string {
	return consts.CodeCooldownPrefix + identifier
}

// GetCodeFailCountKey 获取验证码失败次数在Redis中的键
func GetCodeFailCountKey(identifier string) string {
	return consts.CodeFailCountPrefix + identifier
}

// GetCodeFreezeKey 获取账号冻结在Redis中的键
func GetFreezeKey(identifier string) string {
	return consts.CodeFreezePrefix + identifier
}

// GetLoginFailEmailKey 获取登录失败邮箱计数在Redis中的键
func GetLoginFailEmailKey(email string) string {
	return consts.LoginFailEmailPrefix + email
}

// GetLoginFailIPKey 获取登录失败IP计数在Redis中的键
func GetLoginFailIPKey(ip string) string {
	return consts.LoginFailIPPrefix + ip
}

// GetLoginLockEmailKey 获取登录锁定邮箱在Redis中的键
func GetLoginLockEmailKey(email string) string {
	return consts.LoginLockEmailPrefix + email
}

// GetLoginLockIPKey 获取登录锁定IP在Redis中的键
func GetLoginLockIPKey(ip string) string {
	return consts.LoginLockIPPrefix + ip
}

// SetVerificationCode 存储验证码到Redis
func SetVerificationCode(ctx context.Context, identifier, code string) error {
	key := GetCodeRedisKey(identifier)
	return SetWithExpire(ctx, key, code, time.Duration(consts.CodeExpire)*time.Second)
}

// GetVerificationCode 从Redis获取验证码
func GetVerificationCode(ctx context.Context, identifier string) (string, error) {
	key := GetCodeRedisKey(identifier)
	return Get(ctx, key)
}

// DeleteVerificationCode 从Redis删除验证码
func DeleteVerificationCode(ctx context.Context, identifier string) error {
	key := GetCodeRedisKey(identifier)
	return Del(ctx, key)
}

// CheckCodeCooldown 检查验证码发送是否在冷却期
// 返回是否可以发送、冷却剩余时间（秒）
func CheckCodeCooldown(ctx context.Context, identifier string) (bool, int, error) {
	// 获取冷却键
	cooldownKey := GetCodeCooldownKey(identifier)

	// 查询是否存在冷却记录
	_, err := Get(ctx, cooldownKey)
	if err != nil {
		if IsRedisNil(err) {
			// 不存在冷却记录，可以发送
			return true, 0, nil
		}
		return false, 0, err
	}

	// 获取剩余时间
	remainTime, err := Ttl(ctx, cooldownKey)
	if err != nil {
		return false, 0, err
	}

	// 返回冷却状态和剩余时间
	return false, int(remainTime.Seconds()), nil
}

// SetCodeCooldown 设置验证码冷却期
func SetCodeCooldown(ctx context.Context, identifier string) error {
	// 获取冷却键
	cooldownKey := GetCodeCooldownKey(identifier)

	// 查询是否存在冷却记录
	_, err := Get(ctx, cooldownKey)
	if err != nil && !IsRedisNil(err) {
		return err
	}

	// 设置冷却时间
	var cooldownTime int
	if err != nil && IsRedisNil(err) {
		// 不存在记录，设置首次冷却时间
		cooldownTime = consts.CodeFirstCooldown
	} else {
		// 存在记录，设置第二次冷却时间
		cooldownTime = consts.CodeSecondCooldown
	}

	// 设置冷却时间
	return SetWithExpire(ctx, cooldownKey, "1", time.Duration(cooldownTime)*time.Second)
}

// IncreaseCodeFailCount 增加验证码失败次数
func IncreaseCodeFailCount(ctx context.Context, identifier string) (int, error) {
	// 获取失败次数键
	failCountKey := GetCodeFailCountKey(identifier)

	// 增加失败次数
	count, err := Incr(ctx, failCountKey)
	if err != nil {
		return 0, err
	}

	// 设置过期时间
	err = Expire(ctx, failCountKey, time.Duration(consts.CodeFreezeTime)*time.Second)
	if err != nil {
		// 非致命错误，仅记录日志
		fmt.Println("设置验证码失败次数过期时间失败:", err)
	}

	return int(count), nil
}

// FreezeAccount 冻结账号发送验证码权限
func FreezeAccount(ctx context.Context, identifier string) error {
	// 获取账号冻结键
	freezeKey := GetFreezeKey(identifier)
	// 设置账号冻结
	return SetWithExpire(ctx, freezeKey, "1", time.Duration(consts.CodeFreezeTime)*time.Second)
}

// IsAccountFrozen 检查账号是否被冻结
func IsAccountFrozen(ctx context.Context, identifier string) (bool, error) {
	// 获取账号冻结键
	freezeKey := GetFreezeKey(identifier)
	// 查询是否存在冻结记录
	_, err := Get(ctx, freezeKey)
	if err != nil {
		if IsRedisNil(err) {
			// 不存在冻结记录，账号未被冻结
			return false, nil
		}
		return false, err
	}
	// 存在冻结记录，账号已被冻结
	return true, nil
}

// ResetCodeFailCount 重置验证码失败次数
func ResetCodeFailCount(ctx context.Context, identifier string) error {
	// 获取失败次数键
	failCountKey := GetCodeFailCountKey(identifier)
	// 删除失败次数记录
	return Del(ctx, failCountKey)
}

// IncreaseLoginFailEmailCount 增加登录失败次数（按邮箱）
func IncreaseLoginFailEmailCount(ctx context.Context, email string) (int, error) {
	key := GetLoginFailEmailKey(email)

	// 检查是否存在失败次数记录
	exists, err := Exists(ctx, key)
	if err != nil {
		return 0, err
	}

	var count int

	if !exists {
		// 第一次失败
		err = SetWithExpire(ctx, key, "1", time.Duration(consts.LoginFailExpire)*time.Second)
		count = 1
	} else {
		// 获取当前失败次数
		countStr, err := Get(ctx, key)
		if err != nil {
			return 0, err
		}

		count, err = strconv.Atoi(countStr)
		if err != nil {
			return 0, err
		}

		// 增加失败次数
		count++
		err = SetWithExpire(ctx, key, strconv.Itoa(count), time.Duration(consts.LoginFailExpire)*time.Second)
	}

	if err != nil {
		return 0, err
	}

	return count, nil
}

// IncreaseLoginFailIPCount 增加登录失败次数（按IP）
func IncreaseLoginFailIPCount(ctx context.Context, ip string) (int, error) {
	key := GetLoginFailIPKey(ip)

	// 检查是否存在失败次数记录
	exists, err := Exists(ctx, key)
	if err != nil {
		return 0, err
	}

	var count int

	if !exists {
		// 第一次失败
		err = SetWithExpire(ctx, key, "1", time.Duration(consts.LoginFailExpire)*time.Second)
		count = 1
	} else {
		// 获取当前失败次数
		countStr, err := Get(ctx, key)
		if err != nil {
			return 0, err
		}

		count, err = strconv.Atoi(countStr)
		if err != nil {
			return 0, err
		}

		// 增加失败次数
		count++
		err = SetWithExpire(ctx, key, strconv.Itoa(count), time.Duration(consts.LoginFailExpire)*time.Second)
	}

	if err != nil {
		return 0, err
	}

	return count, nil
}

// LockLoginByEmail 锁定邮箱登录
func LockLoginByEmail(ctx context.Context, email string) error {
	key := GetLoginLockEmailKey(email)
	return SetWithExpire(ctx, key, "1", time.Duration(consts.LoginLockTime)*time.Second)
}

// LockLoginByIP 锁定IP登录
func LockLoginByIP(ctx context.Context, ip string) error {
	key := GetLoginLockIPKey(ip)
	return SetWithExpire(ctx, key, "1", time.Duration(consts.LoginLockTime)*time.Second)
}

// IsLoginLockedByEmail 检查邮箱是否被锁定登录
func IsLoginLockedByEmail(ctx context.Context, email string) (bool, error) {
	key := GetLoginLockEmailKey(email)
	return Exists(ctx, key)
}

// IsLoginLockedByIP 检查IP是否被锁定登录
func IsLoginLockedByIP(ctx context.Context, ip string) (bool, error) {
	key := GetLoginLockIPKey(ip)
	return Exists(ctx, key)
}

// ResetLoginFailEmailCount 重置邮箱登录失败次数
func ResetLoginFailEmailCount(ctx context.Context, email string) error {
	key := GetLoginFailEmailKey(email)
	return Del(ctx, key)
}

// ResetLoginFailIPCount 重置IP登录失败次数
func ResetLoginFailIPCount(ctx context.Context, ip string) error {
	key := GetLoginFailIPKey(ip)
	return Del(ctx, key)
}

// GetTokenBlacklistKey 获取token黑名单在Redis中的键
func GetTokenBlacklistKey(userID string) string {
	return consts.TokenBlacklistPrefix + userID
}

// AddTokenToBlacklist 将token添加到黑名单
func AddTokenToBlacklist(ctx context.Context, userID string) error {
	key := GetTokenBlacklistKey(userID)
	// 使用当前时间戳作为值，便于后续扩展
	return SetWithExpire(ctx, key, time.Now().Unix(), time.Duration(consts.TokenBlacklistExpire)*time.Second)
}

// IsTokenInBlacklist 检查token是否在黑名单中
func IsTokenInBlacklist(ctx context.Context, userID string) (bool, error) {
	key := GetTokenBlacklistKey(userID)
	return Exists(ctx, key)
}

// GetAccountFreezeRemainTime 获取账号冻结剩余时间（秒）
func GetAccountFreezeRemainTime(ctx context.Context, identifier string) (int, error) {
	// 获取账号冻结键
	freezeKey := GetFreezeKey(identifier)
	// 获取剩余时间
	remainTime, err := Ttl(ctx, freezeKey)
	if err != nil {
		return 0, err
	}
	// 返回剩余时间（秒）
	return int(remainTime.Seconds()), nil
}
