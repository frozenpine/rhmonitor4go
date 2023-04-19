package client

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/frozenpine/msgqueue/channel"
	"github.com/frozenpine/msgqueue/core"
	"github.com/frozenpine/rhmonitor4go/service"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

const (
	defaultTimeout = time.Second * 10
)

type RHRiskClient struct {
	ctx         context.Context
	cancel      context.CancelFunc
	start, stop sync.Once

	conn     *grpc.ClientConn
	instance service.RohonMonitorClient
	timeout  time.Duration

	riskSvr     *service.RiskServer
	apiIdentity string
	riskUser    *service.RiskUser

	cmd          *channel.MemoChannel[*Cli]
	orderFlow    atomic.Pointer[channel.MemoChannel[*service.Order]]
	tradeFlow    atomic.Pointer[channel.MemoChannel[*service.Trade]]
	accountFlow  atomic.Pointer[channel.MemoChannel[*service.Account]]
	positionFlow atomic.Pointer[channel.MemoChannel[*service.Position]]
}

func NewRHRiskClient(
	ctx context.Context,
	conn *grpc.ClientConn,
	timeout time.Duration,
	riskSvr *service.RiskServer,
	riskUser *service.RiskUser,
) (*RHRiskClient, error) {
	if conn == nil {
		return nil, errors.New("invalid grpc conn")
	}

	if riskSvr == nil || riskUser == nil {
		return nil, errors.New("invalid risk server params")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if timeout <= 0 {
		timeout = defaultTimeout
	}

	client := &RHRiskClient{
		conn:     conn,
		instance: service.NewRohonMonitorClient(conn),
		timeout:  timeout,
		riskSvr:  riskSvr,
		riskUser: riskUser,
	}
	client.ctx, client.cancel = context.WithCancel(ctx)
	client.cmd = channel.NewMemoChannel[*Cli](client.ctx, "command", 0)

	return client, nil
}

func (c *RHRiskClient) getTimeoutContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = c.timeout
	}

	return context.WithTimeout(c.ctx, timeout)
}

func (c *RHRiskClient) Start() error {
	err := errors.New("already started")

	c.start.Do(func() {
		deadline, cancel := c.getTimeoutContext(-1)

		var result *service.Result

		if result, err = c.instance.Init(deadline, &service.Request{
			Request: &service.Request_Front{
				Front: c.riskSvr,
			},
		}); err != nil {
			return
		} else {
			c.apiIdentity = result.GetApiIdentity()

			log.Printf("Remote risk api initiated: %s", c.apiIdentity)
		}
		cancel()

		go c.serve()
		err = nil
	})

	return err
}

func (c *RHRiskClient) serve() {
	sub, ch := c.cmd.Subscribe("serve", core.Quick)
	defer c.cmd.UnSubscribe(sub)

	for {
		select {
		case <-c.ctx.Done():
			return
		case cli := <-ch:
			log.Print(cli)
		}
	}
}

func (c *RHRiskClient) Stop() error {
	err := errors.New("already stopped")

	c.stop.Do(func() {
		defer c.cancel()

		c.cmd.Release()

		ctx, cancel := c.getTimeoutContext(-1)

		if _, err = c.instance.Release(ctx, &service.Request{
			ApiIdentity: c.apiIdentity,
		}); err == nil {
			err = nil
		}
		cancel()

		err = c.conn.Close()
	})

	return err
}
