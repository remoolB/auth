package service

import (
	"auth/biz/application/dto/Auth/Practice"
	"auth/biz/infrastructure/consts"
	"auth/biz/infrastructure/email"
	"auth/biz/infrastructure/jwt"
	"auth/biz/infrastructure/mapper/user"
	"auth/biz/infrastructure/util"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// AuthService 身份验证服务接口
type AuthService interface {
	// SendVerificationCode 发送验证码
	SendVerificationCode(ctx context.Context, req *Practice.SendVerificationCodeReq) (*Practice.SendVerificationCodeResp, error)
	// VerifyCode 验证验证码
	VerifyCode(ctx context.Context, req *Practice.VerifyCodeReq) (*Practice.VerifyCodeResp, error)
	// Register 用户注册
	Register(ctx context.Context, req *Practice.RegisterReq) (*Practice.RegisterResp, error)
	// Login 用户登录
	Login(ctx context.Context, req *Practice.LoginReq, clientIP string) (*Practice.LoginResp, error)
	// GetUserInfo 获取用户信息
	GetUserInfo(ctx context.Context, userID string, userEmail string) (*Practice.GetUserInfoResp, error)
	// KickUser 踢出用户
	KickUser(ctx context.Context, req *Practice.KickUserReq, currentUserID string) (*Practice.KickUserResp, error)
}

// AuthServiceImpl 身份验证服务实现
type AuthServiceImpl struct {
	userDAO user.IUserDAO
}

// NewAuthService 创建身份验证服务实例
func NewAuthService() AuthService {
	return &AuthServiceImpl{
		userDAO: user.NewUserDAO(),
	}
}

// SendVerificationCode 发送验证码
func (s *AuthServiceImpl) SendVerificationCode(ctx context.Context, req *Practice.SendVerificationCodeReq) (*Practice.SendVerificationCodeResp, error) {
	// 检查账户是否被冻结
	isFrozen, err := util.IsAccountFrozen(ctx, req.Email)
	if err != nil {
		fmt.Println("检查账户冻结状态失败:", err)
		return &Practice.SendVerificationCodeResp{
			Code:    consts.ErrRedis,
			Msg:     consts.ErrMsg[consts.ErrRedis],
			Message: "检查账户状态失败: " + err.Error(),
		}, err
	}

	if isFrozen {
		return &Practice.SendVerificationCodeResp{
			Code:    consts.ErrAccountFrozen,
			Msg:     consts.ErrMsg[consts.ErrAccountFrozen],
			Message: "账号已被冻结，请稍后再试",
		}, nil
	}

	// 检查发送频率限制
	canSend, remainSeconds, err := util.CheckCodeCooldown(ctx, req.Email)
	if err != nil {
		fmt.Println("检查验证码冷却时间失败:", err)
		return &Practice.SendVerificationCodeResp{
			Code:    consts.ErrRedis,
			Msg:     consts.ErrMsg[consts.ErrRedis],
			Message: "检查验证码发送频率失败: " + err.Error(),
		}, err
	}

	if !canSend {
		return &Practice.SendVerificationCodeResp{
			Code:    consts.ErrCodeTooFrequent,
			Msg:     consts.ErrMsg[consts.ErrCodeTooFrequent],
			Message: fmt.Sprintf("验证码发送过于频繁，请等待%d秒后再试", remainSeconds),
		}, nil
	}

	// 生成验证码
	code := util.GenerateVerificationCode()
	fmt.Println("生成的验证码:", code, "邮箱:", req.Email)

	// 存储验证码到Redis
	redisKey := util.GetCodeRedisKey(req.Email)
	err = util.SetWithExpire(ctx, redisKey, code, time.Duration(consts.CodeExpire)*time.Second)
	if err != nil {
		fmt.Println("Redis存储验证码失败:", err)
		return &Practice.SendVerificationCodeResp{
			Code:    consts.ErrRedis,
			Msg:     consts.ErrMsg[consts.ErrRedis],
			Message: "存储验证码失败: " + err.Error(),
		}, err
	}

	// 设置验证码发送冷却时间
	err = util.SetCodeCooldown(ctx, req.Email)
	if err != nil {
		fmt.Println("设置验证码冷却时间失败:", err)
		// 非致命错误，继续流程
	}

	// 发送验证码邮件
	err = email.SendVerificationCode(req.Email, code)
	if err != nil {
		fmt.Println("发送验证码邮件失败:", err)
		return &Practice.SendVerificationCodeResp{
			Code:    consts.ErrSystem,
			Msg:     consts.ErrMsg[consts.ErrSystem],
			Message: "发送验证码邮件失败: " + err.Error(),
		}, err
	}

	// 返回成功响应
	return &Practice.SendVerificationCodeResp{
		Code:    consts.Success,
		Msg:     "验证码发送成功",
		Message: "验证码已发送到您的邮箱，请查收",
	}, nil
}

// VerifyCode 验证验证码
func (s *AuthServiceImpl) VerifyCode(ctx context.Context, req *Practice.VerifyCodeReq) (*Practice.VerifyCodeResp, error) {
	// 检查账户是否被冻结
	isFrozen, err := util.IsAccountFrozen(ctx, req.Email)
	if err != nil {
		fmt.Println("检查账户冻结状态失败:", err)
		return &Practice.VerifyCodeResp{
			Code:  consts.ErrRedis,
			Msg:   consts.ErrMsg[consts.ErrRedis],
			Valid: false,
		}, err
	}

	if isFrozen {
		return &Practice.VerifyCodeResp{
			Code:  consts.ErrAccountFrozen,
			Msg:   consts.ErrMsg[consts.ErrAccountFrozen],
			Valid: false,
		}, nil
	}

	// 从Redis获取验证码
	redisKey := util.GetCodeRedisKey(req.Email)
	storedCode, err := util.Get(ctx, redisKey)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return &Practice.VerifyCodeResp{
				Code:  consts.ErrVerifyCodeExpired,
				Msg:   consts.ErrMsg[consts.ErrVerifyCodeExpired],
				Valid: false,
			}, nil
		}
		return &Practice.VerifyCodeResp{
			Code:  consts.ErrRedis,
			Msg:   consts.ErrMsg[consts.ErrRedis],
			Valid: false,
		}, err
	}

	// 验证码是否匹配
	valid := storedCode == req.VerifyCode

	// 构建响应
	resp := &Practice.VerifyCodeResp{
		Code:  consts.Success,
		Msg:   "验证成功",
		Valid: valid,
	}

	// 验证失败
	if !valid {
		resp.Code = consts.ErrVerifyCodeInvalid
		resp.Msg = consts.ErrMsg[consts.ErrVerifyCodeInvalid]

		// 增加验证失败次数
		failCount, err := util.IncreaseCodeFailCount(ctx, req.Email)
		if err != nil {
			fmt.Println("增加验证码失败次数出错:", err)
			// 非致命错误，继续流程
		}

		// 如果失败次数达到上限，冻结账号
		if failCount >= consts.CodeMaxFailCount {
			err = util.FreezeAccount(ctx, req.Email)
			if err != nil {
				fmt.Println("冻结账号失败:", err)
				// 非致命错误，继续流程
			}

			resp.Code = consts.ErrAccountFrozen
			resp.Msg = consts.ErrMsg[consts.ErrAccountFrozen]
		}
	} else {
		// 验证成功后删除验证码，防止重复使用
		util.Del(ctx, redisKey)

		// 重置验证失败次数
		util.ResetCodeFailCount(ctx, req.Email)
	}

	return resp, nil
}

// Register 用户注册
func (s *AuthServiceImpl) Register(ctx context.Context, req *Practice.RegisterReq) (*Practice.RegisterResp, error) {
	// 检查账户是否被冻结
	isFrozen, err := util.IsAccountFrozen(ctx, req.Email)
	if err != nil {
		fmt.Println("检查账户冻结状态失败:", err)
		return nil, consts.NewAppErrorWithCode(consts.ErrRedis)
	}

	if isFrozen {
		return nil, consts.NewAppErrorWithCode(consts.ErrAccountFrozen)
	}

	// 验证验证码
	verifyReq := &Practice.VerifyCodeReq{
		Email:      req.Email,
		VerifyCode: req.VerifyCode,
	}
	verifyResp, err := s.VerifyCode(ctx, verifyReq)
	if err != nil {
		return nil, err
	}

	// 验证码无效
	if !verifyResp.Valid {
		// 根据验证响应中的错误码获取对应的错误码常量
		var errCode int
		switch int(verifyResp.Code) {
		case consts.ErrVerifyCodeExpired:
			errCode = consts.ErrVerifyCodeExpired
		case consts.ErrVerifyCodeInvalid:
			errCode = consts.ErrVerifyCodeInvalid
		case consts.ErrAccountFrozen:
			errCode = consts.ErrAccountFrozen
		default:
			errCode = consts.ErrSystem
		}
		return nil, consts.NewAppErrorWithCode(errCode)
	}

	// 用户是否已存在
	mongoCtx, cancel := util.CreateContext()
	defer cancel()

	existingUser, err := s.userDAO.FindByEmail(mongoCtx, req.Email)
	if err != nil {
		return nil, consts.NewAppErrorWithCode(consts.ErrMongo)
	}

	if existingUser != nil {
		return nil, consts.NewAppErrorWithCode(consts.ErrUserAlreadyExist)
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, consts.NewAppErrorWithCode(consts.ErrSystem)
	}

	// 创建用户
	newUser := &user.User{
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	err = s.userDAO.Create(mongoCtx, newUser)
	if err != nil {
		return nil, consts.NewAppErrorWithCode(consts.ErrMongo)
	}

	// 生成JWT令牌
	token, expire, err := jwt.GenerateToken(newUser.ID.Hex(), newUser.Email)
	if err != nil {
		return nil, consts.NewAppErrorWithCode(consts.ErrTokenGenerating)
	}

	// 返回成功响应
	return &Practice.RegisterResp{
		Token:        token,
		AccessExpire: expire,
	}, nil
}

// Login 用户登录
func (s *AuthServiceImpl) Login(ctx context.Context, req *Practice.LoginReq, clientIP string) (*Practice.LoginResp, error) {
	// 检查邮箱是否被锁定
	isEmailLocked, err := util.IsLoginLockedByEmail(ctx, req.Email)
	if err != nil {
		fmt.Println("检查邮箱锁定状态失败:", err)
		return nil, consts.NewAppErrorWithCode(consts.ErrRedis)
	}

	if isEmailLocked {
		fmt.Println("邮箱已被锁定:", req.Email)
		return nil, consts.NewAppErrorWithCode(consts.ErrLoginLocked)
	}

	// 检查IP是否被锁定
	isIPLocked, err := util.IsLoginLockedByIP(ctx, clientIP)
	if err != nil {
		fmt.Println("检查IP锁定状态失败:", err)
		return nil, consts.NewAppErrorWithCode(consts.ErrRedis)
	}

	if isIPLocked {
		fmt.Println("IP已被锁定:", clientIP)
		return nil, consts.NewAppErrorWithCode(consts.ErrLoginLocked)
	}

	// 查找用户
	mongoCtx, cancel := util.CreateContext()
	defer cancel()

	foundUser, err := s.userDAO.FindByEmail(mongoCtx, req.Email)
	if err != nil {
		return nil, consts.NewAppErrorWithCode(consts.ErrMongo)
	}

	// 用户不存在或密码错误时，返回统一的错误信息
	if foundUser == nil {
		// 用户不存在，只增加IP维度的失败次数
		util.HandleLoginFailForNonExistentUser(ctx, clientIP)
		// 返回统一的错误信息：账号或密码错误
		return nil, consts.NewAppErrorWithCode(consts.ErrInvalidCredentials)
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(req.Password))
	if err != nil {
		// 密码错误，同时增加邮箱和IP维度的失败计数
		util.HandleLoginFail(ctx, req.Email, clientIP)
		// 返回统一的错误信息：账号或密码错误
		return nil, consts.NewAppErrorWithCode(consts.ErrInvalidCredentials)
	}

	// 登录成功，重置失败计数
	go func() {
		util.ResetLoginFailEmailCount(context.Background(), req.Email)
		util.ResetLoginFailIPCount(context.Background(), clientIP)
	}()

	// 生成JWT令牌
	token, expire, err := jwt.GenerateToken(foundUser.ID.Hex(), foundUser.Email)
	if err != nil {
		return nil, consts.NewAppErrorWithCode(consts.ErrTokenGenerating)
	}

	// 返回成功响应
	return &Practice.LoginResp{
		AccessToken:  token,
		AccessExpire: expire,
	}, nil
}

// GetUserInfo 获取用户信息
func (s *AuthServiceImpl) GetUserInfo(ctx context.Context, userID string, userEmail string) (*Practice.GetUserInfoResp, error) {
	// 如果没有用户信息，表示未认证
	if userID == "" || userEmail == "" {
		return nil, consts.NewAppErrorWithCode(consts.ErrUnauthorized)
	}

	// 查找用户信息
	mongoCtx, cancel := util.CreateContext()
	defer cancel()

	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, consts.NewAppErrorWithCode(consts.ErrSystem)
	}

	foundUser, err := s.userDAO.FindByID(mongoCtx, id)
	if err != nil {
		return nil, consts.NewAppErrorWithCode(consts.ErrMongo)
	}

	if foundUser == nil {
		return nil, consts.NewAppErrorWithCode(consts.ErrUserNotExist)
	}

	// 将ObjectID转换为int64以兼容现有API
	idAsInt64 := util.ObjectIDToInt64(foundUser.ID)

	// 返回成功响应
	return &Practice.GetUserInfoResp{
		Code:       consts.Success,
		Msg:        "获取用户信息成功",
		Id:         idAsInt64, // 将ObjectID转换为int64
		Email:      foundUser.Email,
		CreateTime: foundUser.CreateTime.Unix(),
	}, nil
}

// KickUser 踢出用户
func (s *AuthServiceImpl) KickUser(ctx context.Context, req *Practice.KickUserReq, currentUserID string) (*Practice.KickUserResp, error) {
	// 验证当前用户是否已认证
	if currentUserID == "" {
		return nil, consts.NewAppErrorWithCode(consts.ErrUnauthorized)
	}

	// 验证管理员权限
	mongoCtx, cancel := util.CreateContext()
	defer cancel()

	// 将当前用户ID转换为ObjectID
	currentUserObjectID, err := primitive.ObjectIDFromHex(currentUserID)
	if err != nil {
		return nil, consts.NewAppErrorWithCode(consts.ErrSystem)
	}

	// 查找当前用户
	isAdmin, err := s.userDAO.CheckIsAdmin(mongoCtx, currentUserObjectID)
	if err != nil {
		return &Practice.KickUserResp{
			Code:    consts.ErrMongo,
			Msg:     consts.ErrMsg[consts.ErrMongo],
			Message: "验证管理员权限失败",
		}, nil
	}

	// 验证管理员权限
	if !isAdmin {
		return &Practice.KickUserResp{
			Code:    consts.ErrPermissionDenied,
			Msg:     consts.ErrMsg[consts.ErrPermissionDenied],
			Message: "您不是管理员，无权执行此操作",
		}, nil
	}

	// 使用int64 ID查找用户
	targetUser, err := s.userDAO.FindByInt64ID(mongoCtx, req.UserId)
	if err != nil {
		return &Practice.KickUserResp{
			Code:    consts.ErrMongo,
			Msg:     consts.ErrMsg[consts.ErrMongo],
			Message: "查询用户失败",
		}, nil
	}

	if targetUser == nil {
		return &Practice.KickUserResp{
			Code:    consts.ErrUserNotExist,
			Msg:     consts.ErrMsg[consts.ErrUserNotExist],
			Message: "用户不存在",
		}, nil
	}

	// 将用户token加入黑名单
	err = util.AddTokenToBlacklist(ctx, targetUser.ID.Hex())
	if err != nil {
		return &Practice.KickUserResp{
			Code:    consts.ErrRedis,
			Msg:     consts.ErrMsg[consts.ErrRedis],
			Message: "将用户token加入黑名单失败",
		}, nil
	}

	// 返回成功响应
	return &Practice.KickUserResp{
		Code:    consts.Success,
		Msg:     "操作成功",
		Message: "用户已被踢出系统，该用户需要重新登录",
	}, nil
}
