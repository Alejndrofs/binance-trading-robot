package expert

import (
	"errors"
	"time"
)

type Balance struct {
	time  time.Time
	value float64
}

type VirtualAccount struct {
	Expert            *ExpertAdvisorCrypto
	asset             string
	balance           []Balance
	firstBalanceIsSet bool
}

func NewVirtualAccount(ea *ExpertAdvisorCrypto, asset string) *VirtualAccount {
	virtualAccount := &VirtualAccount{
		Expert:            ea,
		asset:             asset,
		balance:           make([]Balance, 0),
		firstBalanceIsSet: false,
	}
	current_value, err := virtualAccount.Expert.Account.GetAccountTotalBalanceAs(asset)
	if err == nil {
		virtualAccount.firstBalanceIsSet = true
		balance := Balance{
			time:  time.Now(),
			value: current_value,
		}
		virtualAccount.balance = append(virtualAccount.balance, balance)
	}
	return virtualAccount
}

func (va *VirtualAccount) UpdateBalance() error {
	current_value, err := va.Expert.Account.GetAccountTotalBalanceAs(va.asset)
	if err == nil {
		balance := Balance{
			time:  time.Now(),
			value: current_value,
		}
		va.balance = append(va.balance, balance)
		return nil
	}
	return errors.New("virtual account: update balance: current value not found")
}
