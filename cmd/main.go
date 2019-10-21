package main

import (
	"flag"
	"log"
	"net/http"
	"strings"

	yt "github.com/aurawing/ytttransfer"
	"github.com/aurawing/ytttransfer/eostx"
)

func main() {
	mongoURL := flag.String("mongo-url", "mongodb://127.0.0.1:27017/registry", "MongoDB URL")
	eosURL := flag.String("eos-url", "http://127.0.0.1:8888", "EOS URL")
	snapshot := flag.Bool("snapshot", false, "Whether take a snapshot of EOS balance")
	flag.Parse()

	mongoc, err := yt.NewInstance(*mongoURL)
	if err != nil {
		panic(err.Error())
	}
	etx, err := eostx.NewInstance(*eosURL)
	if err != nil {
		panic(err.Error())
	}
	if snapshot {
		log.Println("Starting take a snapshot of EOS balances...")
		accounts, err := etx.GetAccounts()
		if err != nil {
			panic(err.Error())
		}
		mongoc.Snapshot(accounts)
		return
	}
	http.HandleFunc("/reg", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		account := r.Form.Get("account")
		if strings.Trim(account, " ") == "" {
			w.WriteHeader(400)
			return
		}
		ethaddr := r.Form.Get("ethaddr")
		if strings.Trim(ethaddr, " ") == "" {
			w.WriteHeader(400)
			return
		}
		// pubkey := r.Form.Get("pubkey")
		// if strings.Trim(pubkey, " ") == "" {
		// 	w.WriteHeader(400)
		// 	return
		// }
		sig := r.Form.Get("sig")
		if strings.Trim(sig, " ") == "" {
			w.WriteHeader(400)
			return
		}
	})
	err = http.ListenAndServe(":8000", nil)
	if err != nil {
		panic(err.Error())
	}
}
