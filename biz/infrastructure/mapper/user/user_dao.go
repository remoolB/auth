package user

import (
	"auth/biz/infrastructure/consts"
	"auth/biz/infrastructure/util"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// IUserDAO 用户数据访问接口
type IUserDAO interface {
	// Create 创建用户
	Create(ctx context.Context, user *User) error
	// FindByEmail 通过邮箱查找用户
	FindByEmail(ctx context.Context, email string) (*User, error)
	// FindByID 通过ID查找用户
	FindByID(ctx context.Context, id primitive.ObjectID) (*User, error)
	// FindByTimestamp 通过时间戳查找用户
	FindByTimestamp(ctx context.Context, timestamp time.Time) ([]*User, error)
	// FindByInt64ID 通过int64类型的ID查找用户
	FindByInt64ID(ctx context.Context, id int64) (*User, error)
	// Delete 删除用户
	Delete(ctx context.Context, id primitive.ObjectID) error
	// Update 更新用户
	Update(ctx context.Context, user *User) error
	// 检查用户是否为管理员
	CheckIsAdmin(ctx context.Context, id primitive.ObjectID) (bool, error)
}

// UserDAO MongoDB实现的用户DAO
type UserDAO struct{}

// 确保UserDAO实现了IUserDAO接口
var _ IUserDAO = (*UserDAO)(nil)

// NewUserDAO 创建用户DAO实例
func NewUserDAO() IUserDAO {
	return &UserDAO{}
}

// 获取用户集合
func (d *UserDAO) getCollection() (*mongo.Collection, error) {
	return util.GetCollection(consts.UserCollection)
}

// Create 创建用户
func (d *UserDAO) Create(ctx context.Context, user *User) error {
	// 获取集合
	collection, err := d.getCollection()
	if err != nil {
		return err
	}

	// 设置创建时间
	now := time.Now()
	user.CreateTime = now
	user.UpdateTime = now

	// 设置默认角色
	if user.Role == "" {
		user.Role = consts.RoleUser
	}

	// 插入数据
	_, err = collection.InsertOne(ctx, user)
	return err
}

// FindByEmail 通过邮箱查找用户
func (d *UserDAO) FindByEmail(ctx context.Context, email string) (*User, error) {
	// 获取集合
	collection, err := d.getCollection()
	if err != nil {
		return nil, err
	}

	// 构建查询
	filter := bson.M{"email": email}

	// 执行查询
	var user User
	err = collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // 用户不存在
		}
		return nil, err
	}

	return &user, nil
}

// FindByID 通过ID查找用户
func (d *UserDAO) FindByID(ctx context.Context, id primitive.ObjectID) (*User, error) {
	// 获取集合
	collection, err := d.getCollection()
	if err != nil {
		return nil, err
	}

	// 构建查询
	filter := bson.M{"_id": id}

	// 执行查询
	var user User
	err = collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // 用户不存在
		}
		return nil, err
	}

	return &user, nil
}

// FindByTimestamp 通过时间戳查找用户
func (d *UserDAO) FindByTimestamp(ctx context.Context, timestamp time.Time) ([]*User, error) {
	// 获取集合
	collection, err := d.getCollection()
	if err != nil {
		return nil, err
	}

	// 构建查询，查找创建时间接近该时间戳的用户
	// 由于ObjectID的前4字节包含时间信息，我们可以通过时间范围查询
	startTime := timestamp.Add(-1 * time.Minute)
	endTime := timestamp.Add(1 * time.Minute)

	filter := bson.M{
		"create_time": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	}

	// 执行查询
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// 解析结果
	var users []*User
	err = cursor.All(ctx, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// Delete 删除用户
func (d *UserDAO) Delete(ctx context.Context, id primitive.ObjectID) error {
	// 获取集合
	collection, err := d.getCollection()
	if err != nil {
		return err
	}

	// 构建查询
	filter := bson.M{"_id": id}

	// 执行删除
	_, err = collection.DeleteOne(ctx, filter)
	return err
}

// Update 更新用户
func (d *UserDAO) Update(ctx context.Context, user *User) error {
	// 获取集合
	collection, err := d.getCollection()
	if err != nil {
		return err
	}

	// 设置更新时间
	user.UpdateTime = time.Now()

	// 构建查询
	filter := bson.M{"_id": user.ID}

	// 执行更新
	_, err = collection.ReplaceOne(ctx, filter, user)
	return err
}

// CheckIsAdmin 检查用户是否为管理员
func (d *UserDAO) CheckIsAdmin(ctx context.Context, id primitive.ObjectID) (bool, error) {
	user, err := d.FindByID(ctx, id)
	if err != nil {
		return false, err
	}

	if user == nil {
		return false, nil
	}

	return user.Role == consts.RoleAdmin, nil
}

// FindByInt64ID 通过int64类型的ID查找用户
func (d *UserDAO) FindByInt64ID(ctx context.Context, id int64) (*User, error) {
	// 尝试将int64 ID转换为ObjectID
	objID, err := util.Int64ToObjectID(id)
	if err != nil {
		// 如果无法直接转换，则尝试用时间戳查找，保持向后兼容性
		timestamp := time.Unix(id, 0)
		users, err := d.FindByTimestamp(ctx, timestamp)
		if err != nil {
			return nil, err
		}

		if len(users) == 0 {
			return nil, nil
		}

		// 返回时间戳最接近的用户
		return users[0], nil
	}

	// 使用转换后的ObjectID查找用户
	return d.FindByID(ctx, objID)
}
