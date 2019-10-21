package eostx

import eos "github.com/eoscanada/eos-go"

type AccountsInfo struct {
	Code  string `json:"code"`
	Scope string `json:"scope"`
	Table string `json:"table"`
	Payer string `json:"payer"`
	Count uint64 `json:"count"`
	Bal   int64
}

// YTASymbol represents the standard YTA symbol on the chain.
var YTASymbol = eos.Symbol{Precision: 4, Symbol: "YTA"}
