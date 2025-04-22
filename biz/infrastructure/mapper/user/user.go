package user

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email      string             `bson:"email" json:"email"`
	Password   string             `bson:"password" json:"password"`
	Role       string             `bson:"role" json:"role"` // 用户角色：admin-管理员，user-普通用户
	CreateTime time.Time          `bson:"create_time,omitempty" json:"createTime"`
	UpdateTime time.Time          `bson:"update_time,omitempty" json:"updateTime"`
	DeleteTime time.Time          `bson:"delete_time,omitempty" json:"deleteTime"`
}
