package eostx

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/eoscanada/eos-go"
	_ "github.com/eoscanada/eos-go/system"
	_ "github.com/eoscanada/eos-go/token"
)

// NewInstance create a new eostx instance contans connect url, contract owner and it's private key
func NewInstance(url) *eos.API {
	return eos.New(url)
}

//GetExchangeRate get exchange rate between YTA and storage space
func (api *eos.API) GetAccounts() ([]*AccountsInfo, error) {
	req := eos.GetTableByScopeRequest{
		Code:  "eosio.token",
		Table: "accounts",
		Limit: 100,
	}
	accounts := make([]*AccountsInfo, 0)
	for {
		resp, err := api.GetTableByScope(req)
		if err != nil {
			return nil, fmt.Errorf("get table row failed：get accounts：%s\n", err.Error())
		}
		rows := make([]*AccountsInfo, 0)
		err = json.Unmarshal(resp.Rows, &rows)
		if err != nil {
			return nil, err
		}
		for _, acc := range rows {
			assets, err := api.GetCurrencyBalance(eos.AN(acc.Scope), "YTT", "eosio.token")
			if err != nil {
				return nil, err
			}
			for _, a := range assets {
				if a.Symbol.Symbol == "YTT" {
					acc.Bal = int64(a.Amount)
				}
			}
		}
		accounts = append(accounts, rows...)
		if resp.More == "" {
			break
		} else {
			req.LowerBound = resp.More
		}
	}
	return accounts, nil
}

func (api *eos.API) GetPubKey(account string) (string, error) {
	resp, err := eostx.API.GetAccount(eos.AN(account))
	if err != nil {
		return "", err
	}
	perms := resp.Permissions
	for _, p := range perms {
		p.RequiredAuth
	}
}
