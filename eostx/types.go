package eostx

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/eoscanada/eos-go"
)

const JSONTimeFormat = "2006-01-02T15:04:05"

var symbolRegex = regexp.MustCompile("^[0-9],[A-Z]{1,7}$")

type AccountsInfo struct {
	Code  string `json:"code"`
	Scope string `json:"scope"`
	Table string `json:"table"`
	Payer string `json:"payer"`
	Count uint64 `json:"count"`
	Bal   int64
}

type Name string
type AccountName string
type PermissionName string
type Int64 int64
type JSONFloat64 float64
type M map[string]interface{}

type JSONTime struct {
	time.Time
}

type Asset struct {
	Amount Int64
	Symbol
}

type Symbol struct {
	Precision uint8
	Symbol    string

	// Caching of symbol code if it was computed once
	symbolCode uint64
}

type AccountResourceLimit struct {
	Used      Int64 `json:"used"`
	Available Int64 `json:"available"`
	Max       Int64 `json:"max"`
}

type Permission struct {
	PermName     string    `json:"perm_name"`
	Parent       string    `json:"parent"`
	RequiredAuth Authority `json:"required_auth"`
}

type PermissionLevel struct {
	Actor      AccountName    `json:"actor"`
	Permission PermissionName `json:"permission"`
}

type PermissionLevelWeight struct {
	Permission PermissionLevel `json:"permission"`
	Weight     uint16          `json:"weight"` // weight_type
}

type Authority struct {
	Threshold uint32                  `json:"threshold"`
	Keys      []KeyWeight             `json:"keys,omitempty"`
	Accounts  []PermissionLevelWeight `json:"accounts,omitempty"`
	Waits     []WaitWeight            `json:"waits,omitempty"`
}

type KeyWeight struct {
	PublicKey string `json:"key"`
	Weight    uint16 `json:"weight"` // weight_type
}

type WaitWeight struct {
	WaitSec uint32 `json:"wait_sec"`
	Weight  uint16 `json:"weight"` // weight_type
}

type TotalResources struct {
	Owner     AccountName `json:"owner"`
	NetWeight Asset       `json:"net_weight"`
	CPUWeight Asset       `json:"cpu_weight"`
	RAMBytes  Int64       `json:"ram_bytes"`
}

type DelegatedBandwidth struct {
	From      AccountName `json:"from"`
	To        AccountName `json:"to"`
	NetWeight Asset       `json:"net_weight"`
	CPUWeight Asset       `json:"cpu_weight"`
}

type VoterInfo struct {
	Owner             AccountName   `json:"owner"`
	Proxy             AccountName   `json:"proxy"`
	Producers         []AccountName `json:"producers"`
	Staked            Int64         `json:"staked"`
	LastVoteWeight    JSONFloat64   `json:"last_vote_weight"`
	ProxiedVoteWeight JSONFloat64   `json:"proxied_vote_weight"`
	IsProxy           byte          `json:"is_proxy"`
}

type RefundRequest struct {
	Owner       AccountName `json:"owner"`
	RequestTime JSONTime    `json:"request_time"` //         {"name":"request_time", "type":"time_point_sec"},
	NetAmount   Asset       `json:"net_amount"`
	CPUAmount   Asset       `json:"cpu_amount"`
}

type AccountResp struct {
	AccountName            AccountName          `json:"account_name"`
	Privileged             bool                 `json:"privileged"`
	LastCodeUpdate         JSONTime             `json:"last_code_update"`
	Created                JSONTime             `json:"created"`
	CoreLiquidBalance      Asset                `json:"core_liquid_balance"`
	RAMQuota               Int64                `json:"ram_quota"`
	RAMUsage               Int64                `json:"ram_usage"`
	NetWeight              Int64                `json:"net_weight"`
	CPUWeight              Int64                `json:"cpu_weight"`
	NetLimit               AccountResourceLimit `json:"net_limit"`
	CPULimit               AccountResourceLimit `json:"cpu_limit"`
	Permissions            []Permission         `json:"permissions"`
	TotalResources         TotalResources       `json:"total_resources"`
	SelfDelegatedBandwidth DelegatedBandwidth   `json:"self_delegated_bandwidth"`
	RefundRequest          *RefundRequest       `json:"refund_request"`
	VoterInfo              VoterInfo            `json:"voter_info"`
}

// YTASymbol represents the standard YTA symbol on the chain.
var YTASymbol = eos.Symbol{Precision: 4, Symbol: "YTA"}

func (i Int64) MarshalJSON() (data []byte, err error) {
	if i > 0xffffffff || i < -0xffffffff {
		encodedInt, err := json.Marshal(int64(i))
		if err != nil {
			return nil, err
		}
		data = append([]byte{'"'}, encodedInt...)
		data = append(data, '"')
		return data, nil
	}
	return json.Marshal(int64(i))
}

func (i *Int64) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return errors.New("empty value")
	}

	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}

		val, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}

		*i = Int64(val)

		return nil
	}

	var v int64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	*i = Int64(v)

	return nil
}

func (t JSONTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", t.Format(JSONTimeFormat))), nil
}

func (t *JSONTime) UnmarshalJSON(data []byte) (err error) {
	if string(data) == "null" {
		return nil
	}

	t.Time, err = time.Parse(`"`+JSONTimeFormat+`"`, string(data))
	return err
}

func (f *JSONFloat64) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return errors.New("empty value")
	}

	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}

		val, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}

		*f = JSONFloat64(val)

		return nil
	}

	var fl float64
	if err := json.Unmarshal(data, &fl); err != nil {
		return err
	}

	*f = JSONFloat64(fl)

	return nil
}

func (a Asset) Add(other Asset) Asset {
	if a.Symbol != other.Symbol {
		panic("Add applies only to assets with the same symbol")
	}
	return Asset{Amount: a.Amount + other.Amount, Symbol: a.Symbol}
}

func (a Asset) Sub(other Asset) Asset {
	if a.Symbol != other.Symbol {
		panic("Sub applies only to assets with the same symbol")
	}
	return Asset{Amount: a.Amount - other.Amount, Symbol: a.Symbol}
}

func (a Asset) String() string {
	amt := a.Amount
	if amt < 0 {
		amt = -amt
	}
	strInt := fmt.Sprintf("%d", amt)
	if len(strInt) < int(a.Symbol.Precision+1) {
		// prepend `0` for the difference:
		strInt = strings.Repeat("0", int(a.Symbol.Precision+uint8(1))-len(strInt)) + strInt
	}

	var result string
	if a.Symbol.Precision == 0 {
		result = strInt
	} else {
		result = strInt[:len(strInt)-int(a.Symbol.Precision)] + "." + strInt[len(strInt)-int(a.Symbol.Precision):]
	}
	if a.Amount < 0 {
		result = "-" + result
	}

	return fmt.Sprintf("%s %s", result, a.Symbol.Symbol)
}

func (a *Asset) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	asset, err := NewAsset(s)
	if err != nil {
		return err
	}

	*a = asset

	return nil
}

func (a Asset) MarshalJSON() (data []byte, err error) {
	return json.Marshal(a.String())
}

func NewAsset(in string) (out Asset, err error) {
	sec := strings.SplitN(in, " ", 2)
	if len(sec) != 2 {
		return out, fmt.Errorf("invalid format %q, expected an amount and a currency symbol", in)
	}

	if len(sec[1]) > 7 {
		return out, fmt.Errorf("currency symbol %q too long", sec[1])
	}

	out.Symbol.Symbol = sec[1]
	amount := sec[0]
	amountSec := strings.SplitN(amount, ".", 2)

	if len(amountSec) == 2 {
		out.Symbol.Precision = uint8(len(amountSec[1]))
	}

	val, err := strconv.ParseInt(strings.Replace(amount, ".", "", 1), 10, 64)
	if err != nil {
		return out, err
	}

	out.Amount = Int64(val)

	return
}

func NameToSymbol(name Name) (Symbol, error) {
	symbol := Symbol{}
	value, err := StringToName(string(name))
	if err != nil {
		return symbol, fmt.Errorf("name %s is invalid: %s", name, err)
	}

	symbol.Precision = uint8(value & 0xFF)
	symbol.Symbol = SymbolCode(value >> 8).String()

	return symbol, nil
}

func StringToSymbol(str string) (Symbol, error) {
	symbol := Symbol{}
	if !symbolRegex.MatchString(str) {
		return symbol, fmt.Errorf("%s is not a valid symbol", str)
	}

	precision, _ := strconv.ParseUint(string(str[0]), 10, 8)

	symbol.Precision = uint8(precision)
	symbol.Symbol = str[2:]

	return symbol, nil
}

func (s Symbol) SymbolCode() (SymbolCode, error) {
	if s.symbolCode != 0 {
		return SymbolCode(s.symbolCode), nil
	}

	symbolCode, err := StringToSymbolCode(s.Symbol)
	if err != nil {
		return 0, err
	}

	return SymbolCode(symbolCode), nil
}

func (s Symbol) MustSymbolCode() SymbolCode {
	symbolCode, err := StringToSymbolCode(s.Symbol)
	if err != nil {
		panic("Invalid symbol code " + s.Symbol)
	}

	return symbolCode
}

func (s Symbol) ToUint64() (uint64, error) {
	symbolCode, err := s.SymbolCode()
	if err != nil {
		return 0, fmt.Errorf("symbol %s is not a valid symbol code: %s", s.Symbol, err)
	}

	return uint64(symbolCode)<<8 | uint64(s.Precision), nil
}

func (s Symbol) ToName() (string, error) {
	u, err := s.ToUint64()
	if err != nil {
		return "", err
	}
	return NameToString(u), nil
}

func (s Symbol) String() string {
	return fmt.Sprintf("%d,%s", s.Precision, s.Symbol)
}

type SymbolCode uint64

func NameToSymbolCode(name Name) (SymbolCode, error) {
	value, err := StringToName(string(name))
	if err != nil {
		return 0, fmt.Errorf("name %s is invalid: %s", name, err)
	}

	return SymbolCode(value), nil
}

func StringToSymbolCode(str string) (SymbolCode, error) {
	if len(str) > 7 {
		return 0, fmt.Errorf("string is too long to be a valid symbol_code")
	}

	var symbolCode uint64
	for i := len(str) - 1; i >= 0; i-- {
		if str[i] < 'A' || str[i] > 'Z' {
			return 0, fmt.Errorf("only uppercase letters allowed in symbol_code string")
		}

		symbolCode <<= 8
		symbolCode = symbolCode | uint64(str[i])
	}

	return SymbolCode(symbolCode), nil
}

func (sc SymbolCode) ToName() string {
	return NameToString(uint64(sc))
}

func (sc SymbolCode) String() string {
	builder := strings.Builder{}

	symbolCode := uint64(sc)
	for i := 0; i < 7; i++ {
		if symbolCode == 0 {
			return builder.String()
		}

		builder.WriteByte(byte(symbolCode & 0xFF))
		symbolCode >>= 8
	}

	return builder.String()
}

func StringToName(s string) (val uint64, err error) {
	// ported from the eosio codebase, libraries/chain/include/eosio/chain/name.hpp
	var i uint32
	sLen := uint32(len(s))
	for ; i <= 12; i++ {
		var c uint64
		if i < sLen {
			c = uint64(charToSymbol(s[i]))
		}

		if i < 12 {
			c &= 0x1f
			c <<= 64 - 5*(i+1)
		} else {
			c &= 0x0f
		}

		val |= c
	}

	return
}

func charToSymbol(c byte) byte {
	if c >= 'a' && c <= 'z' {
		return c - 'a' + 6
	}
	if c >= '1' && c <= '5' {
		return c - '1' + 1
	}
	return 0
}

var base32Alphabet = []byte(".12345abcdefghijklmnopqrstuvwxyz")

func NameToString(in uint64) string {
	// ported from libraries/chain/name.cpp in eosio
	a := []byte{'.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.'}

	tmp := in
	i := uint32(0)
	for ; i <= 12; i++ {
		bit := 0x1f
		if i == 0 {
			bit = 0x0f
		}
		c := base32Alphabet[tmp&uint64(bit)]
		a[12-i] = c

		shift := uint(5)
		if i == 0 {
			shift = 4
		}

		tmp >>= shift
	}

	return strings.TrimRight(string(a), ".")
}
