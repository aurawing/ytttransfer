package eostx

import (
	"encoding/json"
	"errors"
	"fmt"

	eos "github.com/eoscanada/eos-go"
	_ "github.com/eoscanada/eos-go/system"
	_ "github.com/eoscanada/eos-go/token"
)

type Eostx struct {
	API *eos.API
}

// NewInstance create a new eostx instance contans connect url, contract owner and it's private key
func NewInstance(url string) *Eostx {
	return &Eostx{eos.New(url)}
}

//GetExchangeRate get exchange rate between YTA and storage space
func (api *Eostx) GetAccounts() ([]*AccountsInfo, error) {
	req := eos.GetTableByScopeRequest{
		Code:  "eosio.token",
		Table: "accounts",
		Limit: 100,
	}
	accounts := make([]*AccountsInfo, 0)
	i := 0
	for {
		resp, err := api.API.GetTableByScope(req)
		if err != nil {
			return nil, fmt.Errorf("get table row failed：get accounts：%s\n", err.Error())
		}
		rows := make([]*AccountsInfo, 0)
		err = json.Unmarshal(resp.Rows, &rows)
		if err != nil {
			return nil, err
		}
		for _, acc := range rows {
			i++
			assets, err := api.API.GetCurrencyBalance(eos.AN(acc.Scope), "YTT", "eosio.token")
			if err != nil {
				return nil, err
			}
			for _, a := range assets {
				if a.Symbol.Symbol == "YTT" {
					acc.Bal = int64(a.Amount)
				}
			}
			fmt.Printf("#%d# Fetch account info: %s -> %d\n", i, acc.Scope, acc.Bal)
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

func (api *Eostx) GetPubKey(account string) (string, error) {
	resp := new(AccountResp)
	err := api.API.Call("chain", "get_account", M{"account_name": eos.AN(account)}, resp)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	// resp, err := api.API.GetAccount(eos.AN(account))
	// if err != nil {
	// 	return "", err
	// }
	perms := resp.Permissions
	for _, p := range perms {
		if p.PermName == "active" {
			return p.RequiredAuth.Keys[0].PublicKey, nil
		}
	}
	return "", errors.New("no valid public key")
}
