package monapi

import "github.com/xgfone/ship/v5"

type RouteBinder interface {
	BindRoute(*ship.RouteGroupBuilder)
}
