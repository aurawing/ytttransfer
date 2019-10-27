package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	yt "github.com/aurawing/ytttransfer"
	"github.com/aurawing/ytttransfer/eostx"
	ytcrypto "github.com/yottachain/YTCrypto"
)

type Req struct {
	Account string `json:"account"`
	EthAddr string `json:"ethaddr"`
	Sig     string `json:"sig"`
}

func main() {
	mongoURL := flag.String("mongo-url", "mongodb://127.0.0.1:27017", "MongoDB URL")
	eosURL := flag.String("eos-url", "http://129.28.188.167:8888", "EOS URL")
	snapshot := flag.Bool("snapshot", true, "Take a snapshot of EOS balance")
	port := flag.Int("port", 8080, "Listening port")
	daemon := flag.Bool("d", false, "Run as registry server")
	flag.Parse()

	mgc, err := yt.NewInstance(*mongoURL)
	if err != nil {
		panic(err.Error())
	}
	etx := eostx.NewInstance(*eosURL)

	if *snapshot {
		log.Println("Starting take a snapshot of EOS balances...")
		accounts, err := etx.GetAccounts()
		if err != nil {
			panic(err.Error())
		}
		mgc.Snapshot(accounts, etx)
		return
	}

	if *daemon {
		http.HandleFunc("/balance", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json;charset=UTF-8")
			vals := r.URL.Query()
			if vals == nil || len(vals) == 0 || vals["account"] == nil || len(vals["account"]) == 0 || strings.TrimSpace(vals["account"][0]) == "" {
				w.Write([]byte(formatJson(400, 0, "账号不能为空")))
				return
			}
			account := vals["account"][0]
			reg, err := mgc.GetAccountInfo(account)
			if err != nil {
				if strings.Contains(err.Error(), "no documents in result") || strings.Contains(err.Error(), "resource not found") {
					pubkey, err := etx.GetPubKey(account)
					if err != nil {
						w.Write([]byte(formatJson(400, 0, err.Error())))
						fmt.Printf("!!! balance -> get account info error: %s\n", "账号不存在")
						return
					}
					//balance, _ := etx.GetBalance(account)
					err = mgc.AddRegistry(account, strings.TrimLeft(pubkey, "YTA"), 0)
					if err != nil {
						w.Write([]byte(formatJson(400, 0, err.Error())))
						fmt.Printf("!!! balance -> get account info error: %s\n", "账号不存在")
						return
					}
					w.Write([]byte(formatJson(0, 0, "请求成功")))
					return
				} else {
					w.Write([]byte(formatJson(400, 0, err.Error())))
					fmt.Printf("!!! balance -> get account info error: %s\n", err.Error())
					return
				}
			}
			if reg.Exclude {
				w.Write([]byte(formatJson(400, 0, "账号不存在")))
				return
			}
			w.Write([]byte(formatJson(0, reg.Balance, "请求成功")))
			return
		})

		http.HandleFunc("/ethaddr", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json;charset=UTF-8")
			vals := r.URL.Query()
			if vals == nil || len(vals) == 0 || vals["account"] == nil || len(vals["account"]) == 0 || strings.TrimSpace(vals["account"][0]) == "" {
				w.Write([]byte(formatJson(400, 0, "账号不能为空")))
				return
			}
			account := vals["account"][0]
			reg, err := mgc.GetAccountInfo(account)
			if err != nil {
				w.Write([]byte(formatJson(400, 0, err.Error())))
				fmt.Printf("!!! ethaddr -> get account info error: %s\n", err.Error())
				return
			}
			w.Write([]byte(formatJson2(0, reg.EthAddr, "请求成功")))
			return
		})

		http.HandleFunc("/reg", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json;charset=UTF-8")
			s, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.Write([]byte(formatJson(500, 0, err.Error())))
				return
			}
			formData := new(Req)
			err = json.NewDecoder(bytes.NewReader(s)).Decode(&formData)
			if err != nil {
				w.Write([]byte(formatJson(400, 0, "参数格式不正确")))
				return
			}
			account := formData.Account
			if strings.Trim(account, " ") == "" {
				w.Write([]byte(formatJson(400, 0, "账号不能为空")))
				return
			}
			ethaddr := formData.EthAddr
			if strings.Trim(ethaddr, " ") == "" {
				w.Write([]byte(formatJson(400, 0, "ERC20钱包地址不能为空")))
				return
			}
			sig := formData.Sig
			if strings.Trim(sig, " ") == "" {
				w.Write([]byte(formatJson(400, 0, "签名不能为空")))
				return
			}
			reg, err := mgc.GetAccountInfo(account)
			if err != nil {
				w.Write([]byte(formatJson(500, 0, err.Error())))
				fmt.Printf("!!! reg -> get account info error: %s\n", err.Error())
				return
			}
			pubkey := reg.Pubkey
			if ok := ytcrypto.Verify(pubkey, []byte(fmt.Sprintf("account=%s&ethaddr=%s", account, ethaddr)), sig); ok {
				err = mgc.RegEthAddr(account, ethaddr)
				if err != nil {
					w.Write([]byte(formatJson(500, 0, err.Error())))
					fmt.Printf("!!! reg -> RegEthAddr error: %s\n", err.Error())
					return
				}
				w.Write([]byte(formatJson(0, 0, "ERC20地址注册成功")))
				fmt.Printf("register eth address success: %s -> %s\n", account, ethaddr)
				return
			} else {
				w.Write([]byte(formatJson(401, 0, "签名验证失败")))
				fmt.Printf("!!! reg -> RegEthAddr error: %s\n", "签名验证失败")
				return
			}
		})
		log.Printf(fmt.Sprintf("Server is listening on port %d\n", *port))
		err = http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
		if err != nil {
			panic(err.Error())
		}
		return
	}
	flag.PrintDefaults()
}

func formatJson(code int, data int64, msg string) string {
	return fmt.Sprintf("{\"code\":%d, \"data\": %d, \"msg\":\"%s\"}", code, data, msg)
}

func formatJson2(code int, data string, msg string) string {
	return fmt.Sprintf("{\"code\":%d, \"data\": \"%s\", \"msg\":\"%s\"}", code, data, msg)
}
