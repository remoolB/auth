package jwt

import (
	"auth/biz/infrastructure/config"
	"auth/biz/infrastructure/consts"
	"auth/biz/infrastructure/util"
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Claims 定义JWT的Claims
type Claims struct {
	UserId string `json:"userId"` // 使用string类型与MongoDB的ObjectID兼容
	Email  string `json:"email"`  // 添加邮箱
	jwt.StandardClaims
}

// GenerateToken 生成JWT Token
func GenerateToken(userId string, email string) (string, int64, error) {
	// 获取配置
	jwtConfig := config.GetConfig().JWT

	// 设置过期时间
	expireTime := time.Now().Add(time.Duration(jwtConfig.ExpireTime) * time.Second)
	claims := Claims{
		UserId: userId,
		Email:  email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			NotBefore: time.Now().Unix(),
		},
	}

	// 生成Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtConfig.Secret))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expireTime.Unix(), nil
}

// ParseToken 解析JWT Token
func ParseToken(tokenString string) (*Claims, error) {
	// 获取配置
	jwtConfig := config.GetConfig().JWT

	// 解析Token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtConfig.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证Token
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// 检查token是否在黑名单中
		inBlacklist, err := checkTokenBlacklist(claims.UserId)
		if err != nil {
			return nil, err
		}

		if inBlacklist {
			return nil, errors.New(consts.ErrMsg[consts.ErrTokenBlacklist])
		}

		return claims, nil
	}

	return nil, errors.New(consts.ErrMsg[consts.ErrTokenInvalid])
}

// 检查token是否在黑名单中
func checkTokenBlacklist(userID string) (bool, error) {
	ctx := context.Background()
	return util.IsTokenInBlacklist(ctx, userID)
}
