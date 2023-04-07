package hub

import (
	"sync/atomic"

	"github.com/pkg/errors"
	"google.golang.org/grpc/peer"
)

var (
	ErrClientNotFound = errors.New("client not found")
	ErrPeerNotFound   = errors.New("client peer info not found")
)

type client struct {
	peer  *peer.Peer
	api   *grpcRiskApi
	login atomic.Bool
}

func (c *client) checkConnect() error {
	if c.api.isConnected() {
		return nil
	}

	return errors.New("[grpc] please wait for connected")
}

func (c *client) checkLogin() error {
	if c.login.Load() && c.api.isLoggedIn() {
		return nil
	}

	c.login.Store(false)

	return errors.New("[grpc] please login first")
}
