package ytttransfer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ecrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type TokenTransaction struct {
	client          *ethclient.Client
	contractAddress string
}

func InitTranns(url, contractAddress string) *TokenTransaction {
	rpcDial, err := rpc.Dial(url)
	if err != nil {
		panic(err)
	}
	client := ethclient.NewClient(rpcDial)
	return &TokenTransaction{client: client, contractAddress: contractAddress}
}

func (s *TokenTransaction) Balance(address string) (err error) {
	token, err := NewToken(common.HexToAddress(s.contractAddress), s.client)
	if err != nil {
		return
	}
	bal, err := token.BalanceOf(&bind.CallOpts{}, common.HexToAddress(address))
	fmt.Printf("%d\n", bal)
	return err
}

func (s *TokenTransaction) Transaction(toAddress, privatekey string, tokenAmount float64) (err error) {
	// i, err := ioutil.ReadFile(keyfile)
	// if err != nil {
	// 	return
	// }

	//auth, err := bind.NewTransactor(strings.NewReader(string(i)), pwd)

	privateKey, err := ecrypto.HexToECDSA(privatekey)
	if err != nil {
		panic(err.Error())
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := ecrypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := s.client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	gasPrice, err := s.client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	auth := bind.NewKeyedTransactor(privateKey)
	if err != nil {
		return
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(200000) // in units
	auth.GasPrice = gasPrice

	token, err := NewToken(common.HexToAddress(s.contractAddress), s.client)
	if err != nil {
		return
	}

	amount := big.NewFloat(tokenAmount)
	tenDecimal := big.NewFloat(math.Pow(10, float64(18)))
	convertAmount, _ := new(big.Float).Mul(tenDecimal, amount).Int(&big.Int{})
	auth.GasLimit = 20000000
	txs, err := token.Transfer(auth, common.HexToAddress(toAddress), convertAmount)
	if err != nil {
		return
	}

	fmt.Println("chainId---->", txs.ChainId())
	return
}
