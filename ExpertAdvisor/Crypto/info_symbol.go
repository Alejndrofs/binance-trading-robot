package expert

import (
	"errors"
	"fmt"
	"math"
	"time"
)

type SymbolInfo struct {
	Expert      *ExpertAdvisorCrypto
	SymbolName  string
	Lot         float64
	Synthetic   bool
	BaseAsset   *Asset
	ProfAsset   *Asset
	MinLot      float64
	MinNotional float64
	LotStep     float64
}

func NewSymbol(expert *ExpertAdvisorCrypto, name string, base *Asset, prof *Asset, minLot float64, minNotional float64, lotStep float64) (*SymbolInfo, error) {
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

func NewSyntheticSymbol(expert *ExpertAdvisorCrypto, name string, base *Asset, profit *Asset) (*SymbolInfo, error) {
	newSymbol := &SymbolInfo{
		Expert:     expert,
		SymbolName: name,
		Synthetic:  true,
		BaseAsset:  base,
		ProfAsset:  profit,
	}
	return newSymbol, nil
}

func (s *SymbolInfo) Buy() (bool,error) {
	currentTime := time.Now().String()
	err := s.UpdateLotSize()
	if err != nil {
		return false, errors.New("buy error: "+err.Error())
	}
	fmt.Println(BUY+"\n"+s.SymbolName, "lot ==", s.Lot, s.BaseAsset.AssetName)
	price, err := s.Expert.Exchange.ConvertFactor(s.BaseAsset.AssetName,s.ProfAsset.AssetName)
	if err != nil {
		s.Expert.UpdateErrorsLog("buy error: convert factor error: "+err.Error())
		return false, errors.New("buy error: convert factor error:"+err.Error())
	}
	_, err = s.Expert.TradeManager.ConvertFactorVolume(s.Lot*price, s.ProfAsset.AssetName, s.BaseAsset.AssetName)
	if err != nil {
		s.Expert.UpdateOrdersLog(">>> ORDER_TYPE_BUY: " + currentTime + "\n>>> " + s.SymbolName + "\n>>> ERROR: " + err.Error())
		return false, err
	} else {
		s.Expert.UpdateOrdersLog(">>> ORDER_TYPE_BUY: " + currentTime + "\n>>> " + s.SymbolName + "\n>>> SUCCES")
		return true, nil
	}
}

func (s *SymbolInfo) Sell() (bool,error) {
	currentTime := time.Now().String()
	err := s.UpdateLotSize()
	if err != nil {
		return false, errors.New("sell error: "+err.Error())
	}
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

func (s *SymbolInfo) UpdateLotSize() error {
	if s.Expert.MoneyManagemenType == MONEY_MANAGEMENT_FIX {
		conv, err := s.Expert.Exchange.ConvertFactor(s.Expert.MainAsset, s.BaseAsset.AssetName)
		if err != nil {
			return errors.New("update lot size error: "+err.Error())
		}
		s.Lot = s.Expert.FixTradeVolume * conv
	} else if s.Expert.MoneyManagemenType == MONEY_MANAGEMENT_DYNAMIC {
		accountValue, err := s.Expert.Account.GetAccountTotalBalanceAs(s.Expert.MainAsset)
		if err != nil {
			return errors.New("update lot size error: "+err.Error())
		}
		mainAssetVoluem := s.Expert.DynamicTradePercentage * accountValue
		conv, _ := s.Expert.Exchange.ConvertFactor(s.Expert.MainAsset, s.BaseAsset.AssetName)
		if err != nil {
			return errors.New("update lot size error: "+err.Error())
		}
		s.Lot = mainAssetVoluem * conv
	}
	return nil 
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
			lotA := (ea.MinNotional / lastPrice)
			lotB, err := ea.NormalizeLot(lotA)
			if err != nil {
				return lotA, nil
			}
			return lotB, nil
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
