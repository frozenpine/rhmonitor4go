package hub

import (
	"bytes"
	"context"
	"strconv"
	"sync/atomic"

	"github.com/pkg/errors"
	"google.golang.org/grpc/peer"
)

var (
	ErrClientNotFound = errors.New("client not found")
	ErrPeerNotFound   = errors.New("client peer info not found")
)

type client struct {
	ctx   context.Context
	peer  *peer.Peer
	api   *grpcRiskApi
	login atomic.Bool
	// orderStream
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

func (c *client) disconnect() {
	// c.ctx.
}

func (c *client) String() string {
	result := bytes.NewBuffer(nil)

	result.WriteString(c.peer.Addr.Network())
	result.WriteString("://")
	result.WriteString(c.peer.Addr.String())
	result.WriteString("@")
	result.WriteString(c.api.front.ServerAddr)
	result.WriteString(":")
	result.WriteString(strconv.Itoa(int(c.api.front.ServerPort)))

	if c.login.Load() {
		result.WriteString(": Logged in")
	} else {
		result.WriteString(": Not login")
	}

	return result.String()
}
