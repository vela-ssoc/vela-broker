package mlink

import "github.com/vela-ssoc/backend-common/spdy"

type connect struct {
	ident Ident
	issue Issue
	mux   spdy.Muxer
}
