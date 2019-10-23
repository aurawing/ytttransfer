package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	yt "github.com/aurawing/ytttransfer"
	"github.com/aurawing/ytttransfer/eostx"
	ytcrypto "github.com/yottachain/YTCrypto"
)

func main() {
	mongoURL := flag.String("mongo-url", "mongodb://127.0.0.1:27017", "MongoDB URL")
	eosURL := flag.String("eos-url", "http://129.28.188.167:8888", "EOS URL")
	snapshot := flag.Bool("snapshot", false, "Take a snapshot of EOS balance")
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
			vals := r.URL.Query()
			if vals == nil || len(vals) == 0 || vals["account"] == nil || len(vals["account"]) == 0 || strings.TrimSpace(vals["account"][0]) == "" {
				w.Write([]byte(formatJson(400, 0, "账号不能为空")))
				return
			}
			account := vals["account"][0]
			reg, err := mgc.GetAccountInfo(account)
			if err != nil {
				if strings.Contains(err.Error(), "no documents in result") {
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

		http.HandleFunc("/reg", func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			account := r.Form.Get("account")
			if strings.Trim(account, " ") == "" {
				w.Write([]byte(formatJson(400, 0, "账号不能为空")))
				return
			}
			ethaddr := r.Form.Get("ethaddr")
			if strings.Trim(ethaddr, " ") == "" {
				w.Write([]byte(formatJson(400, 0, "ERC20钱包地址不能为空")))
				return
			}
			sig := r.Form.Get("sig")
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
			if ok := ytcrypto.Verify(pubkey, []byte(fmt.Sprintf("%s#%s", account, ethaddr)), sig); ok {
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
