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
	snapshot := flag.Bool("snapshot", false, "Whether take a snapshot of EOS balance")
	flag.Parse()

	mgc, err := yt.NewInstance(*mongoURL)
	if err != nil {
		panic(err.Error())
	}
	etx := eostx.NewInstance(*eosURL)

	// _, _ = etx.GetPubKey("zzzzzzzzzzz1")

	if *snapshot {
		log.Println("Starting take a snapshot of EOS balances...")
		accounts, err := etx.GetAccounts()
		if err != nil {
			panic(err.Error())
		}
		mgc.Snapshot(accounts, etx)
		return
	}

	http.HandleFunc("/balance", func(w http.ResponseWriter, r *http.Request) {
		vals := r.URL.Query()
		if vals == nil || len(vals) == 0 || vals["account"] == nil || len(vals["account"]) == 0 || strings.TrimSpace(vals["account"][0]) == "" {
			w.WriteHeader(400)
			w.Write([]byte("账号不能为空"))
			return
		}
		account := vals["account"][0]
		reg, err := mgc.GetAccountInfo(account)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(fmt.Sprintf("错误：%s", err.Error())))
			return
		}
		if reg.Exclude {
			w.WriteHeader(400)
			w.Write([]byte(fmt.Sprintf("账号不存在")))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf("%d", reg.Balance)))
	})

	http.HandleFunc("/reg", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		account := r.Form.Get("account")
		if strings.Trim(account, " ") == "" {
			w.WriteHeader(400)
			w.Write([]byte("账号不能为空"))
			return
		}
		ethaddr := r.Form.Get("ethaddr")
		if strings.Trim(ethaddr, " ") == "" {
			w.WriteHeader(400)
			w.Write([]byte("ERC20钱包地址不能为空"))
			return
		}
		sig := r.Form.Get("sig")
		if strings.Trim(sig, " ") == "" {
			w.WriteHeader(400)
			w.Write([]byte("签名不能为空"))
			return
		}
		reg, err := mgc.GetAccountInfo(account)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(fmt.Sprintf("错误：%s", err.Error())))
			return
		}
		pubkey := reg.Pubkey
		if ok := ytcrypto.Verify(pubkey, []byte(fmt.Sprintf("%s#%s", account, ethaddr)), sig); ok {
			w.WriteHeader(200)
			return
		} else {
			w.WriteHeader(401)
			w.Write([]byte("签名验证失败"))
			return
		}
	})

	err = http.ListenAndServe(":8000", nil)
	if err != nil {
		panic(err.Error())
	}
}
