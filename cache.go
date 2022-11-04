package rhmonitor4go

type InvestorCache struct {
	data    map[string]*Investor
	ordered []string
}

func (cache *InvestorCache) Size() int {
	return len(cache.data)
}

func (cache *InvestorCache) AddInvestor(investor *Investor) string {
	identity := investor.Identity()

	cache.ordered = append(cache.ordered, identity)
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

func (cache *InvestorCache) OrderedForEach(fn func(string, *Investor) bool) {
	for _, identity := range cache.ordered {
		if investor, exist := cache.data[identity]; exist {
			if !fn(identity, investor) {
				break
			}
		}
	}
}

type AccountCache struct {
	data    map[string]*Account
	ordered []string
}

func (cache *AccountCache) Size() int {
	return len(cache.data)
}

func (cache *AccountCache) AddAccount(acct *Account) string {
	identity := acct.Identity()

	cache.ordered = append(cache.ordered, identity)
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

func (cache *AccountCache) OrderedForEach(fn func(string, *Account) bool) {
	for _, identity := range cache.ordered {
		if account, exist := cache.data[identity]; exist {
			if !fn(identity, account) {
				break
			}
		}
	}
}
