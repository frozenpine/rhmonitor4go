package client

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/frozenpine/rhmonitor4go/service"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
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
}

func NewRHRiskClient(
	ctx context.Context,
	conn *grpc.ClientConn,
	timeout int,
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

	client := &RHRiskClient{}

	client.ctx, client.cancel = context.WithCancel(ctx)
	client.conn = conn
	client.instance = service.NewRohonMonitorClient(conn)
	client.timeout = time.Duration(timeout)

	client.riskSvr = riskSvr
	client.riskUser = riskUser

	return client, nil
}

func (c *RHRiskClient) getTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.ctx, c.timeout)
}

func (c *RHRiskClient) Start() error {
	err := errors.New("already started")

	c.start.Do(func() {
		deadline, cancel := c.getTimeoutContext()

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
	for {
		select {
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *RHRiskClient) Stop() error {
	err := errors.New("already stopped")

	c.stop.Do(func() {
		_, err = c.instance.Release(c.ctx, &service.Request{
			ApiIdentity: c.apiIdentity,
		})

		c.cancel()

		err = nil
	})

	return err
}
