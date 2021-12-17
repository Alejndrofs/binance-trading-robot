package futures

import (
	"errors"
	"fmt"
	"math"
	"time"
)

type SymbolInfo struct {
	Expert      *ExpertAdvisorFutures
	SymbolName  string
	Lot         float64
	Synthetic   bool
	BaseAsset   *Asset
	ProfAsset   *Asset
	MinLot      float64
	MinNotional float64
	LotStep     float64
}

func NewSymbol(expert *ExpertAdvisorFutures, name string, base *Asset, prof *Asset, minLot float64, minNotional float64, lotStep float64) (*SymbolInfo, error) {
	newSymbol := &SymbolInfo{
		Expert:      expert,
		SymbolName:  name,
		Synthetic:   false,
		BaseAsset:   base,
		ProfAsset:   prof,
		MinLot:      minLot,
		MinNotional: minNotional,
		LotStep:     lotStep,
	}
	return newSymbol, nil
}

func NewSyntheticSymbol(expert *ExpertAdvisorFutures, name string, base *Asset, profit *Asset) (*SymbolInfo, error) {
	newSymbol := &SymbolInfo{
		Expert:     expert,
		SymbolName: name,
		Synthetic:  true,
		BaseAsset:  base,
		ProfAsset:  profit,
	}
	return newSymbol, nil
}

func (s *SymbolInfo) Buy(price float64) (succes bool, err error) {
	currentTime := time.Now().String()
	s.UpdateLotSize()
	fmt.Println(BUY+"\n"+s.SymbolName, "lot ==", s.Lot, s.BaseAsset.AssetName)
	_, err = s.Expert.TradeManager.ConvertFactorVolume(s.Lot*price, s.ProfAsset.AssetName, s.BaseAsset.AssetName)
	if err != nil {
		s.Expert.UpdateOrdersLog(">>> ORDER_TYPE_BUY: " + currentTime + "\n>>> " + s.SymbolName + "\n>>> ERROR: " + err.Error())
		return false, err
	} else {
		s.Expert.UpdateOrdersLog(">>> ORDER_TYPE_BUY: " + currentTime + "\n>>> " + s.SymbolName + "\n>>> SUCCES")
		return true, nil
	}
}

func (s *SymbolInfo) Sell(price float64) (succes bool, err error) {
	currentTime := time.Now().String()
	s.UpdateLotSize()
	fmt.Println(SELL+"\n"+s.SymbolName, "lot ==", s.Lot, s.BaseAsset.AssetName)
	_, err = s.Expert.TradeManager.ConvertFactorVolume(s.Lot, s.BaseAsset.AssetName, s.ProfAsset.AssetName)
	if err != nil {
		s.Expert.UpdateOrdersLog(">>> ORDER_TYPE_SELL: " + currentTime + "\n>>> " + s.SymbolName + "\n>>> ERROR: " + err.Error())
		return false, err
	} else {
		s.Expert.UpdateOrdersLog(">>> ORDER_TYPE_SELL: " + currentTime + "\n>>> " + s.SymbolName + "\n>>> SUCCES")
		return true, nil
	}
}

func (s *SymbolInfo) UpdateLotSize() {
	if s.Expert.MoneyManagemenType == MONEY_MANAGEMENT_FIX {
		conv, _ := s.Expert.Exchange.ConvertFactor(s.Expert.MainAsset, s.BaseAsset.AssetName)
		s.Lot = s.Expert.FixTradeVolume * conv
	} else if s.Expert.MoneyManagemenType == MONEY_MANAGEMENT_DYNAMIC {
		accountValue, _ := s.Expert.Account.GetAccountTotalBalanceAs(s.Expert.MainAsset)
		mainAssetVoluem := s.Expert.DynamicTradePercentage * accountValue
		conv, _ := s.Expert.Exchange.ConvertFactor(s.Expert.MainAsset, s.BaseAsset.AssetName)
		s.Lot = mainAssetVoluem * conv
	}
}

func (ea *SymbolInfo) GetMinimumLotSize() (float64, error) {
	lastPrice, err := ea.Expert.Exchange.ConvertFactor(ea.BaseAsset.AssetName, ea.ProfAsset.AssetName)
	if err != nil {
		ea.Expert.UpdateErrorsLog("get minimum lot size error: " + err.Error())
		return 0, err
	}
	if ea.MinNotional != 0 {
		lot := (ea.MinNotional / lastPrice)
		return math.Max(lot, ea.MinLot), nil
	} else {
		ea.MinNotional, err = ea.Expert.Exchange.SymbolInfoDouble(ea.SymbolName, SYMBOL_INFO_MIN_NOTIONAL)
		if err != nil {
			ea.Expert.UpdateErrorsLog("get minimum lot size error: " + err.Error())
			return 0, err
		} else {
			lot := (ea.MinNotional / lastPrice)
			return math.Max(lot, ea.MinLot), nil
		}
	}
}

func (ea *SymbolInfo) NormalizeLot(lotsize float64) (lot float64, err error) {
	min, err := ea.GetMinimumLotSize()
	if err != nil {
		ea.Expert.UpdateErrorsLog("normalize lot error: " + err.Error())
		return 0, errors.New("error normalize lot size")
	}
	lot = 0
	for lot <= lotsize || lot <= min {
		lot += ea.LotStep
	}
	return lot, err
}
