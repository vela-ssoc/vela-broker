package route

import "github.com/xgfone/ship/v5"

// RegRouter 注册路由
type RegRouter interface {
	// RegRoute 添加路由
	RegRoute(*ship.RouteGroupBuilder)
}
