# 身份验证服务 (Auth Service)

这是一个基于MongoDB的身份验证服务，使用Go语言和Hertz框架开发。

## 功能特性

- 用户注册
- 邮箱验证码发送与验证
- 用户登录
- 获取用户信息
- 踢出用户（管理员功能）
- 验证码发送频率限制和防刷机制
- 登录失败限制和账号保护机制

## 技术栈

- Go 1.20+
- Hertz框架 - 高性能API框架
- MongoDB - 用户数据存储
- Redis - 验证码临时存储
- JWT - 用户身份验证
- bcrypt - 密码加密

## 系统架构

项目采用领域驱动设计(DDD)架构理念，分为以下几层：

- **基础设施层** (infrastructure) - 包含配置、工具类、数据库连接等
- **数据访问层** (mapper) - 定义数据访问对象，负责数据的持久化
- **领域层** (domain) - 业务实体和业务逻辑
- **应用层** (application) - 协调领域层与适配器层
- **适配器层** (adaptor) - 处理HTTP请求和响应

## 项目结构

```
├── biz/                                 - 业务逻辑相关目录
│   ├── adaptor/                         - 适配器层（控制器、路由和中间件）
│   │   ├── common.go                    - 通用函数和结构体
│   │   ├── controller/                  - 控制器目录
│   │   │   ├── ping.go                  - 健康检查控制器
│   │   │   └── Practice/                - 实践模块控制器
│   │   │       └── auth_service.go      - 身份验证服务控制器
│   │   ├── middleware/                  - 中间件目录
│   │   │   └── jwt.go                   - JWT验证中间件
│   │   └── router/                      - 路由目录
│   │       ├── register.go              - 路由注册入口
│   │       └── Practice/                - 实践模块路由
│   │           ├── practice.go          - 身份验证服务路由定义
│   │           └── middleware.go        - 路由中间件配置
│   ├── application/                     - 应用层（服务、DTO）
│   │   ├── service/                     - 服务层目录
│   │   │   └── auth.go                  - 身份验证服务实现
│   │   └── dto/                         - 数据传输对象目录
│   │       └── Auth/                    - 身份验证相关DTO
│   │           └── Practice/            - 实践模块DTO
│   │               ├── practice.pb.go   - 身份验证服务协议缓冲
│   │               └── common.pb.go     - 通用数据结构协议缓冲
│   └── infrastructure/                  - 基础设施层
│       ├── config/                      - 配置目录
│       │   └── config.go                - 配置加载与管理
│       ├── consts/                      - 常量定义目录
│       │   ├── consts.go                - 系统常量定义
│       │   └── errors.go                - 错误码和错误信息定义
│       ├── email/                       - 邮件服务目录
│       │   └── email.go                 - 邮件发送实现
│       ├── jwt/                         - JWT工具目录
│       │   └── jwt.go                   - JWT生成和验证
│       ├── mapper/                      - 数据访问对象目录
│       │   └── user/                    - 用户数据访问
│       │       ├── user.go              - 用户实体定义
│       │       └── user_dao.go          - 用户数据访问方法
│       └── util/                        - 工具类目录
│           ├── mongodb.go               - MongoDB连接和操作工具
│           ├── redis.go                 - Redis连接和操作工具
│           ├── verification.go          - 验证码生成与验证工具
│           ├── login_security.go        - 登录安全相关工具
│           └── object_id.go             - ObjectID处理工具
├── main.go                              - 程序入口
├── router.go                            - 路由初始化
├── router_gen.go                        - 自动生成的路由代码
├── go.mod                               - Go模块依赖定义
├── go.sum                               - Go模块依赖校验和
├── .hz                                  - Hertz框架配置
├── Makefile                             - 项目构建脚本
└── .gitignore                           - Git忽略配置
```

## 接口说明

### 1. 发送验证码

- **URL**: `/api/auth/send-code`
- **方法**: `POST`
- **请求参数**:
  ```json
  {
    "email": "user@example.com"
  }
  ```
- **响应**:
  ```json
  {
    "code": 0,
    "msg": "验证码发送成功",
    "message": "验证码已发送到您的邮箱，请查收"
  }
  ```

**频率限制规则**:
- 第一次发送验证码后，需要等待30秒才能再次发送
- 第二次发送验证码后，需要等待60秒才能再次发送
- 如果验证码验证多次失败（5次），账号将被冻结30分钟

**可能的错误码**:
- 2007: 验证码发送过于频繁 - 需要等待冷却时间
- 2008: 账号已被冻结 - 多次验证失败导致暂时无法发送验证码

### 2. 验证验证码

- **URL**: `/api/auth/verify-code`
- **方法**: `POST`
- **请求参数**:
  ```json
  {
    "email": "user@example.com",
    "verifyCode": "123456"
  }
  ```
- **响应**:
  ```json
  {
    "code": 0,
    "msg": "验证成功",
    "valid": true
  }
  ```

**注意事项**:
- 验证码验证连续失败5次后，账号将被冻结30分钟，期间无法发送或验证验证码
- 验证成功后，验证码会被立即删除，不可重复使用

### 3. 用户注册

- **URL**: `/api/auth/register`
- **方法**: `POST`
- **请求参数**:
  ```json
  {
    "email": "user@example.com",
    "password": "password123",
    "verifyCode": "123456"
  }
  ```
- **响应**:
  ```json
  {
    "token": "eyJhbGciOiJ...",
    "accessExpire": 1627894400
  }
  ```

### 4. 用户登录

- **URL**: `/api/auth/login`
- **方法**: `POST`
- **请求参数**:
  ```json
  {
    "email": "user@example.com",
    "password": "password123"
  }
  ```
- **响应**:
  ```json
  {
    "accessToken": "eyJhbGciOiJ...",
    "accessExpire": 1627894400
  }
  ```

**登录失败限制规则**:
- 系统同时跟踪邮箱和IP地址两个维度的登录失败次数
- 为保护用户隐私，所有登录失败（无论是账号不存在还是密码错误）都统一返回"账号或密码错误"的提示
- 针对不存在的账号尝试登录，系统只会增加IP维度的失败计数，避免暴露账号是否存在的信息
- 针对已存在的账号登录失败（密码错误），系统会同时增加邮箱和IP两个维度的失败计数
- 任一维度连续失败5次后，账号或IP将被锁定30分钟
- 登录成功后，相应邮箱和IP的失败计数会被重置
- 锁定期间，无法通过该邮箱或IP地址登录系统

**可能的错误码**:
- 2010: 账号或密码错误 - 统一的错误提示，不区分账号不存在或密码错误
- 2009: 登录已被锁定 - 多次登录失败导致暂时无法登录

### 5. 获取用户信息

- **URL**: `/api/auth/user-info`
- **方法**: `GET`
- **请求头**: 
  ```
  Authorization: Bearer eyJhbGciOiJ...
  ```
- **响应**:
  ```json
  {
    "code": 0,
    "msg": "获取用户信息成功",
    "id": 1627894400,
    "email": "user@example.com",
    "createTime": 1627808000
  }
  ```

### 6. 踢出用户（管理员功能）

- **URL**: `/api/auth/kick`
- **方法**: `POST`
- **请求头**: 
  ```
  Authorization: Bearer eyJhbGciOiJ...  // 管理员token
  ```
- **请求参数**:
  ```json
  {
    "userId": 1627894400
  }
  ```
- **响应**:
  ```json
  {
    "code": 0,
    "msg": "操作成功",
    "message": "用户已被踢出系统，该用户需要重新登录"
  }
  ```

**功能说明**：
- 此接口只允许管理员用户访问，系统会检查当前用户是否具有管理员权限
- 踢出用户后，该用户的token会被加入黑名单，使其无法继续访问需要认证的接口
- 被踢出的用户需要重新登录才能继续使用系统
- userId 参数是由MongoDB的ObjectID转换而来的唯一标识符（int64格式）
- 系统内部会将此int64标识符转换回MongoDB ObjectID或通过创建时间查找用户




