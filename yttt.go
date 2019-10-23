package ytttransfer

import (
	"context"
	"log"
	"strings"

	"github.com/aurawing/ytttransfer/eostx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Registry struct {
	Account string `json:"_id"`
	Pubkey  string `json:"pubkey"`
	Balance int64  `json:"balance"`
	EthAddr string `json:"ethaddr"`
	Exclude bool   `json:"exclude"`
}

type Mongoc struct {
	Client *mongo.Client
}

func NewInstance(mongoURL string) (*Mongoc, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURL))
	if err != nil {
		return nil, err
	}
	return &Mongoc{client}, nil
}

func (client *Mongoc) Snapshot(accounts []*eostx.AccountsInfo, etx *eostx.Eostx) {
	for i, acc := range accounts {
		pubkey, err := etx.GetPubKey(acc.Scope)
		if err != nil {
			log.Printf("#%d# !!! get pubkey failed: %s ,error: %s\n", acc.Scope, err.Error())
			pubkey = ""
		} else {
			pubkey = strings.TrimLeft(pubkey, "YTA")
		}
		//_, err = collection.InsertOne(context.Background(), bson.M{"_id": acc.Scope, "pubkey": pubkey, "balance": acc.Bal, "ethaddr": "", "exclude": false})
		err = client.AddRegistry(acc.Scope, pubkey, acc.Bal)
		if err != nil {
			log.Printf("#%d# !!! error when snapshot: %s -> %d\n", i, acc.Scope, acc.Bal)
			log.Printf("    %s\n", err.Error())
		}
		log.Printf("#%d# register account: %s -> %d\n", i, acc.Scope, acc.Bal)
	}
}

func (client *Mongoc) RegEthAddr(account, ethaddr string) error {
	collection := client.Client.Database("ytttransfer").Collection("registry")
	_, err := collection.UpdateOne(context.Background(), bson.M{"_id": account}, bson.M{"$set": bson.M{"ethaddr": ethaddr}})
	if err != nil {
		log.Printf("!!! error when registering ethaddress: %s -> %s\n", account, ethaddr)
		return err
	}
	return nil
}

func (client *Mongoc) AddRegistry(account, pubkey string, balance int64) error {
	collection := client.Client.Database("ytttransfer").Collection("registry")
	_, err := collection.InsertOne(context.Background(), bson.M{"_id": account, "pubkey": pubkey, "balance": balance, "ethaddr": "", "exclude": false})
	return err
}

func (client *Mongoc) GetAccountInfo(account string) (*Registry, error) {
	collection := client.Client.Database("ytttransfer").Collection("registry")
	reg := new(Registry)
	err := collection.FindOne(context.Background(), bson.M{"_id": account}).Decode(&reg)
	if err != nil {
		log.Printf("!!! error when query balance of: %s\n", account)
		return nil, err
	}
	return reg, nil
}
