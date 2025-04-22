package util

import (
	"auth/biz/infrastructure/config"
	"auth/biz/infrastructure/consts"
	"context"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient *mongo.Client
	mongoOnce   sync.Once
	mongoLock   sync.Mutex
)

// GetMongoClient 获取MongoDB客户端连接
func GetMongoClient() (*mongo.Client, error) {
	mongoOnce.Do(func() {
		// 获取配置
		conf := config.GetConfig().MongoDB
		
		// 设置连接选项
		clientOptions := options.Client().ApplyURI(conf.URI)
		
		// 如果有用户名和密码，设置认证
		if conf.Username != "" && conf.Password != "" {
			clientOptions.SetAuth(options.Credential{
				Username: conf.Username,
				Password: conf.Password,
			})
		}
		
		// 设置连接超时
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(consts.MongoTimeout)*time.Second)
		defer cancel()
		
		// 连接到MongoDB
		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			panic(err)
		}
		
		// 检查连接
		err = client.Ping(ctx, nil)
		if err != nil {
			panic(err)
		}
		
		mongoClient = client
	})
	
	return mongoClient, nil
}

// GetCollection 获取指定集合
func GetCollection(collectionName string) (*mongo.Collection, error) {
	client, err := GetMongoClient()
	if err != nil {
		return nil, err
	}
	
	conf := config.GetConfig().MongoDB
	return client.Database(conf.Database).Collection(collectionName), nil
}

// CreateContext 创建带超时的上下文
func CreateContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(consts.MongoTimeout)*time.Second)
} 