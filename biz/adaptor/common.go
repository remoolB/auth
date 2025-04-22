package adaptor

import (
	"auth/biz/infrastructure/consts"
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
	hertz "github.com/cloudwego/hertz/pkg/protocol/consts"
)

// HeaderProvider 请求头提供者
type HeaderProvider struct {
	headers *protocol.ResponseHeader
}

// Get 获取请求头值
func (m *HeaderProvider) Get(key string) string {
	return m.headers.Get(key)
}

// Set 设置请求头值
func (m *HeaderProvider) Set(key, value string) {
	m.headers.Set(key, value)
}

// Keys 获取所有键
func (m *HeaderProvider) Keys() []string {
	out := make([]string, 0)

	m.headers.VisitAll(func(key, value []byte) {
		out = append(out, string(key))
	})

	return out
}

// ResponseData 统一响应数据结构
type ResponseData struct {
	Code    int64       `json:"code"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// PostProcess 处理响应
func PostProcess(ctx context.Context, c *app.RequestContext, req, resp any, err error) {
	switch err {
	case nil:
		c.JSON(hertz.StatusOK, resp)
	default:
		// 处理错误响应
		if code, ok := err.(consts.ErrorWithCode); ok {
			c.JSON(hertz.StatusOK, ResponseData{
				Code: int64(code.ErrorCode()),
				Msg:  code.Error(),
			})
		} else {
			c.JSON(hertz.StatusInternalServerError, ResponseData{
				Code: consts.ErrSystem,
				Msg:  consts.ErrMsg[consts.ErrSystem],
			})
		}
	}
}
