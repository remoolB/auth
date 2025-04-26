我先查看项目的README和规范文档，了解现有的身份验证系统。

Ran tool
让我检查一下当前的用户认证实现，首先看看基础设施层的相关文件：

Ran tool
现在我来看看现有的用户实体模型：

Read file: biz/infrastructure/mapper/user/user.go
接下来，我来看看用户数据访问对象：

Read file: biz/infrastructure/mapper/user/user_dao.go
现在，让我查看控制器层的代码：

Ran tool

Read file: biz/adaptor/controller/Practice/auth_service.go

Ran tool

Read file: biz/application/service/auth.go
现在我了解了现有系统的实现。根据用户的需求，我来设计一个支持多种登录方式（邮箱和手机号）的认证系统，并确保可扩展性和查询效率。

# 登录认证系统改造设计方案

## 1. 设计思路

根据用户需求，我们需要对现有的认证系统进行改造，使其支持多种登录方式（邮箱和手机号）。用户的想法是将"登录方式"设计为一个单独的表，这是一个很好的思路，符合可扩展性和性能要求。我将基于这个思路进行设计。

### 1.1 核心实体设计

我们将引入以下实体：

1. **User**: 核心用户实体，包含用户的基本信息
2. **IdentityProvider**: 身份提供者实体，关联到用户，包含具体的登录方式（邮箱、手机号等）
3. **VerificationCode**: 验证码实体，与身份提供者关联

### 1.2 数据模型设计

#### User 实体（现有基础上修改）
```go
type User struct {
    ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Username   string             `bson:"username,omitempty" json:"username"`
    Password   string             `bson:"password" json:"password"`
    Role       string             `bson:"role" json:"role"` // 用户角色：admin-管理员，user-普通用户
    CreateTime time.Time          `bson:"create_time,omitempty" json:"createTime"`
    UpdateTime time.Time          `bson:"update_time,omitempty" json:"updateTime"`
    DeleteTime time.Time          `bson:"delete_time,omitempty" json:"deleteTime"`
}
```

#### IdentityProvider 实体（新增）
```go
type IdentityType string

const (
    EmailIdentity    IdentityType = "email"
    PhoneIdentity    IdentityType = "phone"
    // 未来可扩展更多登录方式，如微信、GitHub等
)

type IdentityProvider struct {
    ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    UserID     primitive.ObjectID `bson:"user_id" json:"userId"`
    Type       IdentityType       `bson:"type" json:"type"`
    Identifier string             `bson:"identifier" json:"identifier"` // 邮箱地址或手机号
    Verified   bool               `bson:"verified" json:"verified"`     // 是否已验证
    CreateTime time.Time          `bson:"create_time,omitempty" json:"createTime"`
    UpdateTime time.Time          `bson:"update_time,omitempty" json:"updateTime"`
    DeleteTime time.Time          `bson:"delete_time,omitempty" json:"deleteTime"`
}
```

### 1.3 索引设计

为确保查询效率，我们将在MongoDB中创建以下索引：

1. **User 集合**:
    - `_id`: 主键索引（MongoDB自动创建）

2. **IdentityProvider 集合**:
    - `user_id`: 用于高效查询用户的所有登录方式
    - `type`, `identifier`: 复合索引，用于高效查询特定类型的登录标识符
    - 索引应该有唯一性约束，确保同一类型的登录标识符不会重复

### 1.4 存储结构

MongoDB 集合设计：
1. `users`: 存储用户基本信息
2. `identity_providers`: 存储用户的登录方式信息

## 2. 功能流程设计

### 2.1 验证码发送流程

1. 接收验证码请求（包含登录方式类型和标识符）
2. 根据登录方式类型选择合适的验证码发送通道（邮件或短信）
3. 生成验证码并存储在Redis中，设置过期时间和防刷机制
4. 发送验证码并返回响应

### 2.2 用户注册流程

1. 接收注册请求（包含登录方式类型、标识符、密码和验证码）
2. 验证验证码
3. 检查标识符是否已被注册
4. 创建新用户记录
5. 创建新的身份提供者记录，关联到新用户
6. 生成JWT令牌并返回

### 2.3 用户登录流程

1. 接收登录请求（包含登录方式类型、标识符和密码）
2. 查找对应的身份提供者记录
3. 根据身份提供者找到关联的用户
4. 验证密码
5. 生成JWT令牌并返回

### 2.4 绑定新登录方式流程

1. 接收绑定请求（包含登录方式类型、标识符和验证码）
2. 验证请求的JWT令牌，确认用户身份
3. 验证验证码
4. 检查标识符是否已被其他用户使用
5. 创建新的身份提供者记录，关联到当前用户
6. 返回成功响应

### 2.5 解绑登录方式流程

1. 接收解绑请求（包含要解绑的身份提供者ID）
2. 验证请求的JWT令牌，确认用户身份
3. 检查用户是否至少还有另一种登录方式（防止无法登录）
4. 删除指定的身份提供者记录
5. 返回成功响应

## 3. API接口设计

### 3.1 发送验证码接口

```
POST /api/auth/send-code
```

请求参数：
```json
{
  "identityType": "email", // 或 "phone"
  "identifier": "user@example.com" // 或手机号 "13800138000"
}
```

响应：
```json
{
  "code": 0,
  "msg": "验证码发送成功",
  "message": "验证码已发送到您的邮箱/手机，请查收"
}
```

### 3.2 用户注册接口

```
POST /api/auth/register
```

请求参数：
```json
{
  "identityType": "email", // 或 "phone"
  "identifier": "user@example.com", // 或手机号
  "password": "password123",
  "verifyCode": "123456"
}
```

响应：
```json
{
  "code": 0,
  "msg": "注册成功",
  "accessToken": "eyJhbGciOiJ...",
  "accessExpire": 1627894400
}
```

### 3.3 用户登录接口

```
POST /api/auth/login
```

请求参数：
```json
{
  "identityType": "email", // 或 "phone"
  "identifier": "user@example.com", // 或手机号
  "password": "password123"
}
```

响应：
```json
{
  "code": 0,
  "msg": "登录成功",
  "accessToken": "eyJhbGciOiJ...",
  "accessExpire": 1627894400
}
```

### 3.4 绑定新登录方式接口

```
POST /api/auth/bind-identity
```

请求头：
```
Authorization: Bearer eyJhbGciOiJ...
```

请求参数：
```json
{
  "identityType": "phone", // 要绑定的新登录方式
  "identifier": "13800138000", // 新的标识符
  "verifyCode": "123456"
}
```

响应：
```json
{
  "code": 0,
  "msg": "绑定成功",
  "message": "已成功绑定新的登录方式"
}
```

### 3.5 解绑登录方式接口

```
POST /api/auth/unbind-identity
```

请求头：
```
Authorization: Bearer eyJhbGciOiJ...
```

请求参数：
```json
{
  "identityId": "507f1f77bcf86cd799439011" // 要解绑的身份提供者ID
}
```

响应：
```json
{
  "code": 0,
  "msg": "解绑成功",
  "message": "已成功解绑登录方式"
}
```

### 3.6 获取当前用户的所有登录方式

```
GET /api/auth/identities
```

请求头：
```
Authorization: Bearer eyJhbGciOiJ...
```

响应：
```json
{
  "code": 0,
  "msg": "获取成功",
  "identities": [
    {
      "id": "507f1f77bcf86cd799439011",
      "type": "email",
      "identifier": "user@example.com",
      "verified": true,
      "createTime": 1627808000
    },
    {
      "id": "507f1f77bcf86cd799439012",
      "type": "phone",
      "identifier": "13800138000",
      "verified": true,
      "createTime": 1627808100
    }
  ]
}
```

## 4. 安全设计

### 4.1 验证码安全

1. 为每种登录方式单独实现验证码发送频率限制
2. 根据标识符（邮箱/手机号）跟踪验证失败次数
3. 超过指定次数（如5次）后，冻结该标识符一段时间（如30分钟）

### 4.2 登录安全

1. 跟踪每种登录方式的失败尝试次数
2. 跟踪IP地址的登录失败尝试次数
3. 任一维度达到阈值后锁定相应标识符或IP一段时间
4. 登录成功后重置失败计数

### 4.3 数据安全

1. 密码使用bcrypt算法加密存储
2. 敏感操作（如绑定/解绑登录方式）必须先验证身份
3. 解绑登录方式时确保用户至少保留一种验证过的登录方式

## 5. 扩展性设计

本设计的扩展性体现在：

1. **登录方式扩展**：`IdentityType`枚举可以轻松添加新的登录方式（如社交媒体账号、OAuth等）
2. **验证方式扩展**：可以为不同类型的身份提供者实现不同的验证方式
3. **存储结构扩展**：`IdentityProvider`实体可以添加特定登录方式的额外字段

## 6. 开发步骤规划

1. 创建新的数据模型和DAO层
2. 修改现有服务层以适应新的数据模型
3. 实现新的API接口
4. 更新控制器和路由
5. 实现短信验证码发送服务（新增功能）
6. 编写单元测试
7. 为管理员扩展相关功能

## 7. 总结

这个方案通过引入身份提供者概念，将用户与其登录方式解耦，有效解决了多种登录方式的问题，同时保持了查询效率和可扩展性。通过完善的索引设计，确保了各种登录方式的查询性能一致，无论是邮箱还是手机号都能高效查询。

方案的主要优势：

1. **高效**：通过适当的索引保证查询效率
2. **可扩展**：容易添加新的登录方式
3. **灵活**：用户可以拥有多种登录方式
4. **安全**：维持了现有系统的安全机制，并扩展到新的登录方式

需要注意的实现细节：

1. 为`identity_providers`集合创建合适的索引
2. 扩展验证码发送服务，支持短信验证码
3. 在解绑登录方式时确保用户至少保留一种验证过的登录方式

您认为这个设计方案是否可行？如果您有其他建议或顾虑，我们可以进一步讨论。
