package rhmonitor4go

type AccountCache struct {
	cache map[string]*Account
}

func (cache *AccountCache) AddAccount(acct *Account) {
	cache.cache[acct.AccountID] = acct
}

func (cache *AccountCache) GetAccount(accountID string) *Account {
	return cache.cache[accountID]
}
