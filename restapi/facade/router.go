package facade

import "github.com/xgfone/ship/v5"

type Router interface {
	Route(ship *ship.RouteGroupBuilder)
}
