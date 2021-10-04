package expert

import (
	"context"
	"errors"
)

type ENUM_ACCOUNT_INFO_INT string
type ENUM_ACCOUNT_INFO_BOOL string
type ENUM_ACCOUNT_INFO_BALANCE string

const (
	ACCOUNT_INFO_INT_MAKER_COMMISSION  ENUM_ACCOUNT_INFO_INT     = "MakerCommission"
	ACCOUNT_INFO_INT_TAKER_COMMISSION  ENUM_ACCOUNT_INFO_INT     = "TakerCommission"
	ACCOUNT_INFO_INT_BUYER_COMMISSION  ENUM_ACCOUNT_INFO_INT     = "BuyerCommission"
	ACCOUNT_INFO_INT_SELLER_COMMISSION ENUM_ACCOUNT_INFO_INT     = "SellerCommission"
	ACCOUNT_INFO_BOOL_CAN_TRADE        ENUM_ACCOUNT_INFO_BOOL    = "CanTrade"
	ACCOUNT_INFO_BOOL_CAN_WITHDRAW     ENUM_ACCOUNT_INFO_BOOL    = "CanWithdraw"
	ACCOUNT_INFO_BOOL_CAN_DEPOSIT      ENUM_ACCOUNT_INFO_BOOL    = "CanDeposit"
	ACCOUNT_INFO_FREE_BALANCE          ENUM_ACCOUNT_INFO_BALANCE = "FREE_BALANCE"
	ACCOUNT_INFO_LOCKED_BALANCE        ENUM_ACCOUNT_INFO_BALANCE = "LOCKED_BALANCE"
)

type Asset struct {
	AssetName string
	Symbols   []*SymbolInfo
	Balance   float64
}

func NewAsset(asset string) *Asset {
	newAsset := &Asset{
		AssetName: asset,
		Symbols:   make([]*SymbolInfo, 0),
	}
	return newAsset
}

func (a *Asset) IsEnoughMoney(amount float64) bool {
	if a.Balance > amount {
		return true
	} else {
		return false
	}
}

type AccountInfo struct {
	Expert      *ExpertAdvisorCrypto
	SymbolsList []*SymbolInfo
	AssetsList  []*Asset
}

func NewAccountInfo(expert *ExpertAdvisorCrypto) *AccountInfo {
	AI := &AccountInfo{
		Expert:      expert,
		SymbolsList: make([]*SymbolInfo, 0),
		AssetsList:  make([]*Asset, 0),
	}
	AI.FindAssetList()
	AI.UpdateSymbolsInfo()
	AI.UpdateBalances()
	return AI
}

func (ea *AccountInfo) AssetExists(asset string) bool {
	for _, v := range ea.AssetsList {
		if asset == v.AssetName {
			return true
		}
	}
	return false
}

func (ea *AccountInfo) GetAsset(name string) (*Asset, error) {
	for i := range ea.AssetsList {
		if name == ea.AssetsList[i].AssetName {
			return ea.AssetsList[i], nil
		}
	}
	return &Asset{}, errors.New("asset not found")
}

func (ea *AccountInfo) GetSymbol(assetA *Asset, assetB *Asset) (*SymbolInfo, error) {
	for i := range ea.AssetsList {
		if assetA.AssetName == ea.AssetsList[i].AssetName {
			for j, s := range ea.AssetsList[i].Symbols {
				if s.BaseAsset.AssetName == assetA.AssetName && s.ProfAsset.AssetName == assetB.AssetName {
					return ea.AssetsList[i].Symbols[j], nil
				}
				if s.ProfAsset.AssetName == assetA.AssetName && s.BaseAsset.AssetName == assetB.AssetName {
					return ea.AssetsList[i].Symbols[j], nil
				}
			}
		}
	}
	return &SymbolInfo{}, errors.New("symbol not found")
}

func (ea *AccountInfo) AccountInfoInt(i ENUM_ACCOUNT_INFO_INT) (int64, error) {
	account, err := ea.Expert.Client.NewGetAccountService().Do(context.Background())
	if err != nil {
		ea.Expert.UpdateErrorsLog("get account info int error: get account info error" + err.Error())
		return 0, err
	}
	switch i {
	case "MakerCommission":
		return account.MakerCommission, nil
	case "TakerCommission":
		return account.TakerCommission, nil
	case "BuyerCommission":
		return account.BuyerCommission, nil
	case "SellerCommission":
		return account.SellerCommission, nil
	}
	return 0, errors.New("account info int ACCOUNT_INFO_INT not defined")
}

func (ea *AccountInfo) AccountInfoBool(i ENUM_ACCOUNT_INFO_BOOL) (bool, error) {
	account, err := ea.Expert.Client.NewGetAccountService().Do(context.Background())
	if err != nil {
		ea.Expert.UpdateErrorsLog("get account info bool error: get account info error" + err.Error())
		return false, err
	}
	switch i {
	case "CanTrade":
		return account.CanTrade, nil
	case "CanWithdraw":
		return account.CanWithdraw, nil
	case "CanDeposit":
		return account.CanDeposit, nil
	default:
		return false, nil
	}
}

func (ea *AccountInfo) AccountInfoBalance(asset string, i ENUM_ACCOUNT_INFO_BALANCE) (float64, error) {
	account, err := ea.Expert.Client.NewGetAccountService().Do(context.Background())
	if err != nil {
		ea.Expert.UpdateErrorsLog("account info balance error: get account info error" + err.Error())
		return 0, nil
	}
	for _, x := range account.Balances {
		if asset == x.Asset {
			switch i {
			case "FREE_BALANCE":
				return x.Free, nil
			case "LOCKED_BALANCE":
				return x.Locked, nil
			default:
				return 0, nil
			}
		}
	}
	return 0, nil
}

func (ea *AccountInfo) GetAccountTotalBalanceAs(asset string) (value float64, err error) {
	ea.UpdateBalances()
	var currentBalance float64 = 0
	for i := range ea.AssetsList {
		if ea.AssetsList[i].Balance != 0 {
			convFactor, err := ea.Expert.Exchange.ConvertFactor(ea.AssetsList[i].AssetName, asset)
			if err == nil {
				currentBalance += ea.AssetsList[i].Balance * convFactor
			} else {
				ea.Expert.UpdateErrorsLog("convert balance to: conversion factor not found")
			}
		}
	}
	return currentBalance, nil
}

func (ea *AccountInfo) FindAssetList() {
	exchange, err := ea.Expert.Client.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		ea.Expert.UpdateErrorsLog("find asset list error: " + err.Error())
	}
	for _, s := range exchange.Symbols {
		base := s.BaseAsset
		prof := s.QuoteAsset
		if !ea.AssetExists(base) {
			newAsset := NewAsset(base)
			ea.AssetsList = append(ea.AssetsList, newAsset)
		}
		if !ea.AssetExists(prof) {
			newAsset := NewAsset(prof)
			ea.AssetsList = append(ea.AssetsList, newAsset)
		}
	}
	AllSymbols := exchange.Symbols
	for i := 0; i < len(ea.AssetsList); i++ {
		for j := i + 1; j < len(ea.AssetsList); j++ {
			for x := range AllSymbols {
				base := AllSymbols[x].BaseAsset
				prof := AllSymbols[x].QuoteAsset
				if ea.AssetsList[j].AssetName == base && ea.AssetsList[i].AssetName == prof {
					minLot, _ := ea.Expert.Exchange.SymbolInfoDouble(AllSymbols[x].Symbol, SYMBOL_INFO_MIN_LOT)
					lotStep, _ := ea.Expert.Exchange.SymbolInfoDouble(AllSymbols[x].Symbol, SYMBOL_INFO_STEP_LOT)
					minNotional, _ := ea.Expert.Exchange.SymbolInfoDouble(AllSymbols[x].Symbol, SYMBOL_INFO_MIN_NOTIONAL)
					newSymbol, err := NewSymbol(ea.Expert, AllSymbols[x].Symbol, ea.AssetsList[j], ea.AssetsList[i], minLot, minNotional, lotStep)
					if err != nil {
						ea.Expert.UpdateErrorsLog("find asset list: new symbol error: " + err.Error())
					} else {
						ea.AssetsList[i].Symbols = append(ea.AssetsList[i].Symbols, newSymbol)
						ea.AssetsList[j].Symbols = append(ea.AssetsList[j].Symbols, newSymbol)
						ea.SymbolsList = append(ea.SymbolsList, newSymbol)
						AllSymbols = append(AllSymbols[:x], AllSymbols[x+1:]...)
						break
					}
				} else if ea.AssetsList[i].AssetName == base && ea.AssetsList[j].AssetName == prof {
					minLot, _ := ea.Expert.Exchange.SymbolInfoDouble(AllSymbols[x].Symbol, SYMBOL_INFO_MIN_LOT)
					lotStep, _ := ea.Expert.Exchange.SymbolInfoDouble(AllSymbols[x].Symbol, SYMBOL_INFO_STEP_LOT)
					minNotional, _ := ea.Expert.Exchange.SymbolInfoDouble(AllSymbols[x].Symbol, SYMBOL_INFO_MIN_NOTIONAL)
					newSymbol, err := NewSymbol(ea.Expert, AllSymbols[x].Symbol, ea.AssetsList[i], ea.AssetsList[j], minLot, minNotional, lotStep)
					if err != nil {
						ea.Expert.UpdateErrorsLog("find asset list: new symbol error: " + err.Error())
					} else {
						ea.AssetsList[i].Symbols = append(ea.AssetsList[i].Symbols, newSymbol)
						ea.AssetsList[j].Symbols = append(ea.AssetsList[j].Symbols, newSymbol)
						ea.SymbolsList = append(ea.SymbolsList, newSymbol)
						AllSymbols = append(AllSymbols[:x], AllSymbols[x+1:]...)
						break
					}
				}
			}
		}
	}
}

func (ea *AccountInfo) UpdateSymbolsInfo() error {
	AllSymbols := ea.SymbolsList
	for i := 0; i < len(ea.AssetsList); i++ {
		for j := range ea.AssetsList[i].Symbols {
			for x := range AllSymbols {
				if ea.AssetsList[i].Symbols[j].SymbolName == AllSymbols[x].SymbolName {
					ea.AssetsList[i].Symbols[j].MinLot, _ = ea.Expert.Exchange.SymbolInfoDouble(AllSymbols[x].SymbolName, SYMBOL_INFO_MIN_LOT)
					ea.AssetsList[i].Symbols[j].LotStep, _ = ea.Expert.Exchange.SymbolInfoDouble(AllSymbols[x].SymbolName, SYMBOL_INFO_STEP_LOT)
					ea.AssetsList[i].Symbols[j].MinNotional, _ = ea.Expert.Exchange.SymbolInfoDouble(AllSymbols[x].SymbolName, SYMBOL_INFO_MIN_NOTIONAL)
					break
				}
			}
		}
	}
	return nil
}

func (ea *AccountInfo) UpdateBalances() error {
	var attempts uint = 10
	account, err := ea.Expert.Client.NewGetAccountService().Do(context.Background())
	for err != nil && attempts > 0 {
		account, err = ea.Expert.Client.NewGetAccountService().Do(context.Background())
		attempts--
	}
	if attempts == 0 {
		ea.Expert.UpdateErrorsLog("update balances error: " + err.Error())
		return errors.New("update balances error: " + err.Error())
	}
	for i := range ea.AssetsList {
		for j := range account.Balances {
			if ea.AssetsList[i].AssetName == account.Balances[j].Asset {
				ea.AssetsList[i].Balance = account.Balances[j].Free
			}
		}
	}
	return nil
}
