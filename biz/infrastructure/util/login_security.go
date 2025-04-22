package util

import (
	"auth/biz/infrastructure/consts"
	"context"
	"fmt"
)

// HandleLoginFailForNonExistentUser 处理不存在用户的登录失败，只增加IP维度的失败次数
func HandleLoginFailForNonExistentUser(ctx context.Context, ip string) {
	// 只增加IP失败计数
	ipFailCount, err := IncreaseLoginFailIPCount(ctx, ip)
	if err != nil {
		fmt.Println("增加IP登录失败计数出错:", err)
		return
	}

	fmt.Printf("尝试登录不存在的账号 - IP: %s (失败次数: %d)\n", ip, ipFailCount)

	// 如果IP失败次数达到阈值，锁定IP
	if ipFailCount >= consts.LoginMaxFailCount {
		err = LockLoginByIP(ctx, ip)
		if err != nil {
			fmt.Println("锁定IP登录失败:", err)
		} else {
			fmt.Println("已锁定IP登录:", ip)
		}
	}
}

// HandleLoginFail 处理登录失败，记录失败次数并在达到阈值时锁定账号
func HandleLoginFail(ctx context.Context, email, ip string) {
	// 增加邮箱失败计数
	emailFailCount, err := IncreaseLoginFailEmailCount(ctx, email)
	if err != nil {
		fmt.Println("增加邮箱登录失败计数出错:", err)
		return
	}

	// 增加IP失败计数
	ipFailCount, err := IncreaseLoginFailIPCount(ctx, ip)
	if err != nil {
		fmt.Println("增加IP登录失败计数出错:", err)
		return
	}

	fmt.Printf("登录失败 - 邮箱: %s (失败次数: %d), IP: %s (失败次数: %d)\n",
		email, emailFailCount, ip, ipFailCount)

	// 如果邮箱失败次数达到阈值，锁定邮箱
	if emailFailCount >= consts.LoginMaxFailCount {
		err = LockLoginByEmail(ctx, email)
		if err != nil {
			fmt.Println("锁定邮箱登录失败:", err)
		} else {
			fmt.Println("已锁定邮箱登录:", email)
		}
	}

	// 如果IP失败次数达到阈值，锁定IP
	if ipFailCount >= consts.LoginMaxFailCount {
		err = LockLoginByIP(ctx, ip)
		if err != nil {
			fmt.Println("锁定IP登录失败:", err)
		} else {
			fmt.Println("已锁定IP登录:", ip)
		}
	}
}
