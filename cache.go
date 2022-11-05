package rhmonitor4go

import (
	"sync/atomic"

	"github.com/frozenpine/channel"
)

func waitBoolFlag(flag *atomic.Bool, v bool) <-chan struct{} {
	ch := make(chan struct{})

	go func() {
		for !flag.CompareAndSwap(v, v) {
		}
		ch <- struct{}{}
	}()

	return ch
}

type RequestCache struct {
	isConnected  atomic.Bool
	reqAfterConn []func() int

	isLoggedIn    atomic.Bool
	reqAfterLogin []func() int

	isInvestorReady atomic.Bool
	reqAfterReady   []func() int
}

func (req *RequestCache) WaitConnected() {
	<-waitBoolFlag(&req.isConnected, true)
}

func (req *RequestCache) WaitConnectedAndDo(fn func() int) int {
	req.WaitConnected()

	req.reqAfterConn = append(req.reqAfterConn, fn)

	return fn()
}

func (req *RequestCache) IsConnected() bool {
	return req.isConnected.Load()
}

func (req *RequestCache) SetConnected(v bool) bool {
	return req.isConnected.CompareAndSwap(!v, v)
}

func (req *RequestCache) RedoConnected() (rtn int) {
	for _, fn := range req.reqAfterConn {
		rtn = fn()

		if rtn != 0 {
			break
		}
	}

	return
}

func (req *RequestCache) WaitLogin() {
	<-waitBoolFlag(&req.isLoggedIn, true)
}

func (req *RequestCache) WaitLoginAndDo(fn func() int) int {
	req.WaitLogin()

	req.reqAfterLogin = append(req.reqAfterLogin, fn)

	return fn()
}

func (req *RequestCache) IsLoggedIn() bool {
	return req.isLoggedIn.Load()
}

func (req *RequestCache) SetLogin(v bool) bool {
	return req.isLoggedIn.CompareAndSwap(!v, v)
}

func (req *RequestCache) RedoLoggedIn() (rtn int) {
	for _, fn := range req.reqAfterLogin {
		rtn = fn()

		if rtn != 0 {
			break
		}
	}

	return
}

func (req *RequestCache) WaitInvestorReady() {
	<-waitBoolFlag(&req.isInvestorReady, true)
}

func (req *RequestCache) WaitInvestorReadyAndDo(fn func() int) int {
	req.WaitInvestorReady()

	req.reqAfterReady = append(req.reqAfterReady, fn)

	return fn()
}

func (req *RequestCache) IsInvestorReady() bool {
	return req.isInvestorReady.Load()
}

func (req *RequestCache) SetInvestorReady(v bool) bool {
	return req.isInvestorReady.CompareAndSwap(!v, v)
}

func (req *RequestCache) RedoInvestorReady() (rtn int) {
	for _, fn := range req.reqAfterReady {
		rtn = fn()

		if rtn != 0 {
			break
		}
	}

	return
}

type InvestorCache struct {
	data map[string]*Investor
}

func (cache *InvestorCache) Size() int {
	return len(cache.data)
}

func (cache *InvestorCache) AddInvestor(investor *Investor) string {
	identity := investor.Identity()

	cache.data[identity] = investor

	return identity
}

func (cache *InvestorCache) GetInvestor(identity string) *Investor {
	return cache.data[identity]
}

func (cache *InvestorCache) ForEach(fn func(string, *Investor) bool) {
	for identity, investor := range cache.data {
		if !fn(identity, investor) {
			break
		}
	}
}

type AccountCache struct {
	data     map[string]*Account
	dataChan *channel.Hub[*Account]
}

func (cache *AccountCache) Size() int {
	return len(cache.data)
}

func (cache *AccountCache) AddAccount(acct *Account) string {
	identity := acct.Identity()

	cache.data[identity] = acct

	return identity
}

func (cache *AccountCache) GetAccount(identity string) *Account {
	return cache.data[identity]
}

func (cache *AccountCache) ForEach(fn func(string, *Account) bool) {
	for identity, account := range cache.data {
		if !fn(identity, account) {
			break
		}
	}
}
