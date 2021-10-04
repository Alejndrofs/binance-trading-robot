package expert

type VirtualAccount struct {
	balance []float64
}

func NewVirtualAccount(firstBalance float64) *VirtualAccount {
	virtualAccount := &VirtualAccount{
		balance: make([]float64, 0),
	}
	virtualAccount.balance = append(virtualAccount.balance, firstBalance)
	return virtualAccount
}

func (va *VirtualAccount) UpdateBalance(new_value float64) {
	va.balance = append(va.balance, new_value)
}
