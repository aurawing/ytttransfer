package ytttransfer

import (
	"context"

	"github.com/aurawing/ytttransfer/eostx"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Registry struct {
	Account string `json:"_id"`
	Balance int64 `json:"balance"`
	EthAddr string `json:"ethaddr"`
	Exclude bool `json:"exclude"`
}

func NewInstance(mongoURL string) (*mongo.Client, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURL))
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (client *mongo.Client) Snapshot(accounts []*eostx.AccountsInfo) {
	collection := client.Database("ytttransfer").Collection("registry")
	for i, acc := range accounts {
		_, err := collection.InsertOne(context.Background(), bson.M{"_id": acc.Scope, "balance": acc.Bal, "ethaddr": "", "exclude": false})
		if err != nil {
			log.Printf("#%d# !!! error when snapshot: %s -> %d\n", i, acc.Scope, acc.Bal)
			log.Printf("    %s\n", err.Error())
		}
		log.Printf("#%d# register account: %s -> %d\n", i, acc.Scope, acc.Bal)
	}
}


func (client *mongo.Client) RegEthAddr(account, ethaddr string) error {
	collection := client.Database("ytttransfer").Collection("registry")
	_, err := collection.UpdateOne(context.Background(), bson.M{"_id": account}, bson.M{"$set": bson.M{"ethaddr": ethaddr}})
	if err != nil {
		log.Printf("!!! error when registering ethaddress: %s -> %s\n", account, ethaddr)
		return err
	}
	return nil
}

func (client *mongo.Client) GetBalance(account string) (int64, error) {
	collection := client.Database("ytttransfer").Collection("registry")
	reg := new(Registry)
	err := collection.FindOne(context.Background(), bson.M{"_id": account}).Decode(&reg)
	if err != nil {
		log.Printf("!!! error when query balance of: %s\n", account)
		return 0, err
	}
	return reg.Balance, nil
}