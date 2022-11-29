package data

import (
	"Lealra/config"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

var Db *mongo.Database

func getMongoConn() *mongo.Database {
	dsConfig := config.Settings.DataSource
	var clientOpts *options.ClientOptions
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dsConfig.TimeOut)*time.Second)
	defer cancel()
	if dsConfig.Auth {
		credential := options.Credential{
			Username: dsConfig.Username,
			Password: dsConfig.Password,
		}
		clientOpts = options.Client().ApplyURI("mongodb://localhost:" + dsConfig.Port).SetAuth(credential)
	} else {
		clientOpts = options.Client().ApplyURI("mongodb://localhost:" + dsConfig.Port)
	}
	clientOpts.SetMaxPoolSize(dsConfig.MaxPoolSize)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("已顺利连接至 MongoDB")
	}
	return client.Database(dsConfig.DatabaseName)
}

func ConnInit() {
	Db = getMongoConn()
}
