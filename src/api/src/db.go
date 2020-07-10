package main

import (
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/mongo/readpref"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"

    "context"
    "time"
    "log"
)

func (db *Database) search(category string, filter bson.M) bson.M {
     collection := db.client.Database("Raspisanie").Collection(category)
     var result bson.M

     ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
     defer cancel()

     cur, err := collection.Find(ctx, filter)
     if err != nil {
     	log.Fatal(err)
     }

     for cur.Next(ctx) {	 
   	 err := cur.Decode(&result)
   	 if err != nil {
	    log.Fatal(err)
	 }
     }

     if result != nil {
     	return result["payload"].(primitive.M)
     } else {
        return bson.M{}
     }
}

func (db *Database) pool(category string) map[string] primitive.M {
     var cat_for_search string

     if category == "preps" {
     	cat_for_search = "prep"
     } else if category == "groups" {
        cat_for_search = "group"
     }
     
     collection := db.client.Database("Raspisanie").Collection(category)

     ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
     defer cancel()

     cur, err := collection.Find(ctx, bson.D{})
     if err != nil {
     	log.Fatal(err)
     }
     
     defer cur.Close(ctx)

     output := make(map[string] primitive.M)

     for cur.Next(ctx) {
     	 var result bson.M
	 
   	 err := cur.Decode(&result)
   	 if err != nil {
	    log.Fatal(err)
	 }
	 
	 output[result[cat_for_search].(string)] = result["payload"].(primitive.M)
     }

     if err := cur.Err(); err != nil {
     	log.Fatal(err)
     }

     return output
}

func (db *Database) insert_token(token string) int {
     collection := db.client.Database("db").Collection("tokens")

     ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
     defer cancel()

     itemCount, err := collection.CountDocuments(ctx, bson.M{"token": token})
     if err != nil {
     	log.Panic(err) 	
     }

     if itemCount >= 1 {
     	return -1 
     } else {
	_, err := collection.InsertOne(ctx, bson.M{"token": token})
	if err != nil {
	   log.Panic(err)
	}

     	return 1
     }
}

func (db *Database) check_token(token string) int {
     collection := db.client.Database("db").Collection("tokens")

     ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
     defer cancel()

     itemCount, err := collection.CountDocuments(ctx, bson.M{"token": token})
     if err != nil {
     	log.Panic(err) 	
     }

     if itemCount == 1 {
     	return 1
     } else {
       	return -1
     }
}

func db_init(address string, port string) Database {
     uri := "mongodb://" + address + ":" + port

     ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
     defer cancel()

     client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
     if err != nil {
	panic(err)
     }

     if err := client.Ping(ctx, readpref.Primary()); err != nil {
	panic(err)
     }

     return Database{client}
}

func (db *Database) close() {
     ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
     defer cancel()

     _ = db.client.Disconnect(ctx)
}