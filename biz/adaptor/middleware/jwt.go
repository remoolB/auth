package middleware

import (
	"auth/biz/infrastructure/consts"
	"auth/biz/infrastructure/jwt"
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	hconsts "github.com/cloudwego/hertz/pkg/protocol/consts"
)

// JWTAuth 中间件用于验证用户JWT令牌
func JWTAuth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 从请求头中获取令牌
		authHeader := string(c.Request.Header.Get("Authorization"))
		if authHeader == "" {
			c.JSON(hconsts.StatusUnauthorized, map[string]interface{}{
				"code": consts.ErrUnauthorized,
				"msg":  consts.ErrMsg[consts.ErrUnauthorized],
			})
			c.Abort()
			return
		}

		// 通常令牌格式为 "Bearer {token}"
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(hconsts.StatusUnauthorized, map[string]interface{}{
				"code": consts.ErrTokenInvalid,
				"msg":  consts.ErrMsg[consts.ErrTokenInvalid],
			})
			c.Abort()
			return
		}

		// 验证令牌
		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			c.JSON(hconsts.StatusUnauthorized, map[string]interface{}{
				"code": consts.ErrTokenInvalid,
				"msg":  consts.ErrMsg[consts.ErrTokenInvalid],
			})
			c.Abort()
			return
		}

		// 将用户信息存储在上下文中，便于后续操作
		c.Set("userId", claims.UserId)
		c.Set("userEmail", claims.Email)

		// 继续处理请求
		c.Next(ctx)
	}
}
