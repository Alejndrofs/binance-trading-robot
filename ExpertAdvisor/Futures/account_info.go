package futures

import (
	"errors"
	"context"
	"strconv"
)

type ENUM_ACCOUNT_INFO_INT string
type ENUM_ACCOUNT_INFO_BOOL string
type ENUM_ACCOUNT_INFO_BALANCE string

const (
	ACCOUNT_INFO_INT_FEE_TIER_COMMISSION ENUM_ACCOUNT_INFO_INT   = "FeeTier"
	ACCOUNT_INFO_BOOL_CAN_TRADE        ENUM_ACCOUNT_INFO_BOOL    = "CanTrade"
	ACCOUNT_INFO_BOOL_CAN_WITHDRAW     ENUM_ACCOUNT_INFO_BOOL    = "CanWithdraw"
	ACCOUNT_INFO_BOOL_CAN_DEPOSIT      ENUM_ACCOUNT_INFO_BOOL    = "CanDeposit"
	ACCOUNT_INFO_MARGIN_BALANCE        ENUM_ACCOUNT_INFO_BALANCE = "MarginBalance"
	ACCOUNT_INFO_WALLET_BALANCE        ENUM_ACCOUNT_INFO_BALANCE = "WalletBalance"
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

type AccountInfo struct {
	Expert     *ExpertAdvisorFutures
	SymbolsList	[]*SymbolInfo
	AssetsList 	[]*Asset
}

func NewAccountInfo(expert *ExpertAdvisorFutures) *AccountInfo {
	AI := &AccountInfo{
		Expert: expert,
		SymbolsList: make([]*SymbolInfo, 0),
		AssetsList: make([]*Asset, 0),
	}
	AI.FindAssetList()
	AI.UpdateBalances()
	return AI
}

func (ei *AccountInfo) AssetExists(asset string) bool {
	for _, v := range ei.AssetsList {
		if asset == v.AssetName {
			return true
		}
	}
	return false
}

func (ea *AccountInfo) GetAsset(name string) (*Asset,error) {
	for i := range ea.AssetsList {
		if name == ea.AssetsList[i].AssetName {
			return ea.AssetsList[i], nil
		}
	}
	return &Asset{}, errors.New("asset not found")
}

func (ea *AccountInfo) AccountInfoInt(i ENUM_ACCOUNT_INFO_INT) (int64, error) {
	account, err := ea.Expert.Client.NewGetAccountService().Do(context.Background())
	if err != nil {
		ea.Expert.UpdateErrorsLog("get account info int error: get account info error" + err.Error())
		return 0, err
	}
	switch i {
	case "FeeTier":
		return int64(account.FeeTier), nil
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
	for _, x := range account.Assets {
		if asset == x.Asset {
			switch i {
			case "FREE_BALANCE":
				f, err := strconv.ParseFloat(x.MarginBalance,64)
				if err != nil {
					return 0, errors.New("accoint info balance strconv error")
				}
				return f, nil
			case "LOCKED_BALANCE":
				f, err := strconv.ParseFloat(x.WalletBalance,64)
				if err != nil {
					return 0, errors.New("accoint info balance strconv error")
				}
				return f, nil
			default:
				return 0, nil
			}
		}
	}
	return 0, nil
}

/*
func (ai *AccountInfo) GetAccountInfo() (err error) {
	ai.Account, err = ai.AccountService.Do(context.Background())
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func (ai *AccountInfo) GetAccountSnapshot() (err error) {
	ai.AccountSnapshot, err = ai.AccountSnapshotService.Do(context.Background())
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func (ea *AccountInfo) AccountInfoInt(i ENUM_ACCOUNT_INFO_INT) (int64,error) {
	switch i {
	case "MakerCommission":	return ea.Account.MakerCommission, nil
	case "TakerCommission": return ea.Account.TakerCommission, nil
	case "BuyerCommission":	return ea.Account.BuyerCommission, nil
	case "SellerCommission": return ea.Account.SellerCommission, nil
	default:	return 0, nil
	}
}

func (ea *AccountInfo) AccountInfoBool(i ENUM_ACCOUNT_INFO_BOOL) (bool,error) {
	switch i {
	case "CanTrade": return ea.Account.CanTrade, nil
	case "CanWithdraw": return ea.Account.CanWithdraw, nil
	case "CanDeposit": return ea.Account.CanDeposit, nil
	default: return false, nil
	}
}

func (ea *AccountInfo) AccountInfoBalance(asset string, i ENUM_ACCOUNT_INFO_BALANCE) (float64,error) {
	for _ , x := range ea.Account.Balances {
		if asset == x.Asset {
			switch i {
				case "FREE_BALANCE": return x.Free, nil
				case "LOCKED_BALANCE": return x.Locked, nil
				default: return 0, nil
			}
		}
	}
	return 0, nil
}
*/

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
	account, err := ea.Expert.Client.NewGetAccountService().Do(context.Background())
	if err != nil {
		ei.Expert.UpdateErrorsLog("find asset list error: " + err.Error())
	}
	for _, s := range exchange.Symbols {
		base := s.BaseAsset
		prof := s.QuoteAsset
		if !ei.AssetExists(base) {
			newAsset := NewAsset(base)
			ei.AssetsList = append(ei.AssetsList, newAsset)
		}
		if !ei.AssetExists(prof) {
			newAsset := NewAsset(prof)
			ei.AssetsList = append(ei.AssetsList, newAsset)
		}
	}
	AllSymbols := exchange.Symbols
	for i := 0; i < len(ei.AssetsList); i++ {
		for j := i + 1; j < len(ei.AssetsList); j++ {
			for x := range AllSymbols {
				base := AllSymbols[x].BaseAsset
				prof := AllSymbols[x].QuoteAsset
				if (ei.AssetsList[j].AssetName == base && ei.AssetsList[i].AssetName == prof) {
					var minLot float64
					var lotStep float64
					var minNotional float64
					for _, f := range AllSymbols[x].Filters {
						if f.Type == "LOT_SIZE" {
							minLot = f.MinQty
							lotStep = f.StepSize
						} else if f.Type == "MIN_NOTIONAL" {
							minNotional = f.MinNotional
						}
					}
					newSymbol, err := NewSymbol(ei.Expert, AllSymbols[x].Symbol, ei.AssetsList[j], ei.AssetsList[i], minLot, minNotional, lotStep)
					if err != nil {
						ei.Expert.UpdateErrorsLog("find asset list: new symbol error: " + err.Error())
					} else {
						ei.AssetsList[i].Symbols = append(ei.AssetsList[i].Symbols, newSymbol)
						ei.AssetsList[j].Symbols = append(ei.AssetsList[j].Symbols, newSymbol)
						ei.SymbolsList = append(ei.SymbolsList, newSymbol)
						AllSymbols = append(AllSymbols[:x], AllSymbols[x+1:]...)
						break
					}
				} else if (ei.AssetsList[i].AssetName == base && ei.AssetsList[j].AssetName == prof) {
					var minLot float64
					var lotStep float64
					var minNotional float64
					for _, f := range AllSymbols[x].Filters {
						if f.Type == "LOT_SIZE" {
							minLot = f.MinQty
							lotStep = f.StepSize
						} else if f.Type == "MIN_NOTIONAL" {
							minNotional = f.MinNotional
						}
					}
					newSymbol, err := NewSymbol(ei.Expert, AllSymbols[x].Symbol, ei.AssetsList[i], ei.AssetsList[j], minLot, minNotional, lotStep)
					if err != nil {
						ei.Expert.UpdateErrorsLog("find asset list: new symbol error: " + err.Error())
					} else {
						ei.AssetsList[i].Symbols = append(ei.AssetsList[i].Symbols, newSymbol)
						ei.AssetsList[j].Symbols = append(ei.AssetsList[j].Symbols, newSymbol)
						ei.SymbolsList = append(ei.SymbolsList, newSymbol)
						AllSymbols = append(AllSymbols[:x], AllSymbols[x+1:]...)
						break
					}
				}
			}
		}
	}
}

func (ea *AccountInfo) UpdateSymbolsInfo() error {
	account, err := ea.Expert.Client.NewGetAccountService().Do(context.Background())
	if err != nil {
		ei.Expert.UpdateErrorsLog("update symbols info error: " + err.Error())
		return err
	}
	AllSymbols := exchange.Symbols
	for i := 0; i < len(ei.AssetsList); i++ {
		for j := range ei.AssetsList[i].Symbols {
			for x := range AllSymbols {
				if ei.AssetsList[i].Symbols[j].SymbolName == AllSymbols[x].Symbol {
					for _, f := range AllSymbols[x].Filters {
						if f.Type == "LOT_SIZE" {
							ei.AssetsList[i].Symbols[j].MinLot = f.MinQty
							ei.AssetsList[i].Symbols[j].LotStep = f.StepSize
						} else if f.Type == "MIN_NOTIONAL" {
							ei.AssetsList[i].Symbols[j].MinNotional = f.MinNotional
						}
					}
					break
				}
			}
		}
	}
	return nil
}

func (ea *AccountInfo) UpdateBalances() error {
	account, err := ea.Expert.Client.NewGetAccountService().Do(context.Background())
	var attempts uint = 10
	for err != nil && attempts > 0 {
		account, err = ea.Expert.Client.NewGetAccountService().Do(context.Background())
		attempts--
	}
	if attempts == 0 {
		ea.Expert.UpdateErrorsLog("update balances error: " + err.Error())
		return errors.New("update balances error: " + err.Error())
	}
	for i := range ea.AssetsList {
		for j := range account.Assets {
			if ea.AssetsList[i].AssetName == account.Assets[j].Asset {
				ea.AssetsList[i].Balance = account.Assets[j].WalletBalance
			}
		}
	}
	return nil
}
