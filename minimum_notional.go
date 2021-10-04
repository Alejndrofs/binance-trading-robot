package main

import (
	expert "ExpertAdvisor/Crypto"
	"fmt"
)

func main() {
	fmt.Println("New expert!")

	EA := expert.NewExpertAdvisor("apiKey", "apiSecret")

	fmt.Println("Search symbol with lower min notional!")
	var asset string
	fmt.Println("Type 'exit' to continue")
	for {
		fmt.Print(">>> select asset: ")
		fmt.Scanln(&asset)
		exist := EA.Account.AssetExists(asset)
		if exist {
			break
		} else {
			fmt.Println("Asset not found!")
		}
	}
	for _, s := range EA.Account.SymbolsList {
		if s.BaseAsset.AssetName == asset || s.ProfAsset.AssetName == asset {
			conv, _ := EA.Exchange.ConvertFactor(s.BaseAsset.AssetName, "EUR")
			val, err := s.GetMinimumLotSize()
			if err != nil {
				continue
			}
			val *= conv
			fmt.Println("Min Notional", s.SymbolName, val)		
		}
	}
}
