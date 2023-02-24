package route

import "github.com/xgfone/ship/v5"

type RegRouter interface {
	RegRoute(*ship.RouteGroupBuilder)
}
