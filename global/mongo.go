package global

import (
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"os"
)

var (
	Mongo *mongo.Database
)

func ConnectMongo() {
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGODB_URI"))
	client, err := mongo.Connect(nil, clientOptions)
	if err != nil {
		log.Panic("Trying to connect to MongoDB with the URI " + os.Getenv("MONGODB_URI") + " failed: " + err.Error())
	}
	Mongo = client.Database("main")

	// Check if the connection is alive
	err = Mongo.Client().Ping(nil, nil)
	if err != nil {
		log.Panic("MongoDB connection is not alive: " + err.Error())
	}
}
