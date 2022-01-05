package main

/*
//--- Version 1.0
	-
//--- Version 1.1
	-
//--- Version 1.2
	-
//--- Version 1.3
	- Multisymbolo secuencial
//--- Version 1.4
	- Select by Assets
	- Concurrency at Ontimer
	- Lote dinamico con volumen fijo
	- web socket
//--- Version 1.5
	- Medir variaciones en USDT para normalizar los rangos
//--- Version 1.6
	- Money Management Portfolio Valuation
//---- Version 1.7
	- Lote dinamico con volumen dinamico
	- Sinthetic symbols
//---- Version 1.8
	- Mutiple changes on convert factor & vonvert factor volume
	- Fixed Volume on Queue.Execute to avoid errors
	- Update Symbols info function (min notional,min lot & lot step)
//--- To do
	- Mutiple main currency estrategy: test diferent performance movements
	- RealtimeBacktesting
	- Blockchain database
	- Config file
	- portfolio builder: select grups of assets by separated
	- Corelate portfolios
	- Trade different trangles as a portfolio of strategys
	- Daily reset of the initial price

*/

import (
	expert "ExpertAdvisor/Crypto"
	"fmt"
)

const version float64 = 1.8

var EA *expert.ExpertAdvisorCrypto
var Banner string = " ██████╗ ██████╗ ██╗██████╗ \n██╔════╝ ██╔══██╗██║██╔══██╗\n██║  ███╗██████╔╝██║██║  ██║\n██║   ██║██╔══██╗██║██║  ██║\n╚██████╔╝██║  ██║██║██████╔╝\n ╚═════╝ ╚═╝  ╚═╝╚═╝╚═════╝ \n"

func main() {

	fmt.Println(Banner)
	fmt.Println("Init expert advisor!")
	fmt.Println("Version", version)

	foundAPIkey := false
	foundAPIsecret := false
	var apiKey string
	var apiSecret string

	if !foundAPIkey {
		for {
			fmt.Print(">>> insert API key: ")
			fmt.Scanln(&apiKey)
			if apiKey != "" {
				foundAPIkey = true
			}
			if foundAPIkey {
				break
			} else {
				fmt.Println(">>> empty input!")
			}
		}
	}

	if !foundAPIsecret {
		for {
			fmt.Print(">>> insert API secret: ")
			fmt.Scanln(&apiSecret)
			if apiSecret != "" {
				foundAPIsecret = true
			}
			if foundAPIsecret {
				break
			} else {
				fmt.Println(">>> empty input!")
			}
		}
	}

	EA = expert.NewExpertAdvisor(apiKey, apiSecret)
}
