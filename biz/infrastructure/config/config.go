package config

import (
	"sync"
)

// MongoDB配置
type MongoDBConfig struct {
	URI      string
	Database string
	Username string
	Password string
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// EmailConfig 邮箱配置
type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

// JWT配置
type JWTConfig struct {
	Secret     string
	ExpireTime int64 // token过期时间，单位秒
}

// AppConfig 应用配置
type AppConfig struct {
	MongoDB MongoDBConfig
	Redis   RedisConfig
	Email   EmailConfig
	JWT     JWTConfig
}

// ConfigInstance 单例实例
var instance *AppConfig
var once sync.Once

// GetConfig 获取配置单例
func GetConfig() *AppConfig {
	once.Do(func() {
		instance = &AppConfig{
			MongoDB: MongoDBConfig{
				URI:      "mongodb://localhost:27017",
				Database: "auth_service",
				Username: "",
				Password: "",
			},
			Redis: RedisConfig{
				Host:     "127.0.0.1",
				Port:     6379,
				Password: "ruoyi123",
				DB:       0,
			},
			Email: EmailConfig{
				Host:     "smtp.qq.com",
				Port:     465, // 使用SSL/TLS端口
				Username: "1561662079@qq.com",
				Password: "adxhpprrbbnuiegc",
			},
			JWT: JWTConfig{
				Secret:     " J3w8*Lm!7z@q#P1x",
				ExpireTime: 86400, // 24小时（86400 秒）
			},
		}
	})
	return instance
}
