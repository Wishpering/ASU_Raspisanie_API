package main

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"context"
	"time"
)

func (db *Database) Search(faculty string, group_id string) (bson.M, error) {
	var result bson.M
	collection := db.client.Database("rasp").Collection(faculty)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cur, err := collection.Find(ctx, bson.M{"group": group_id})
	if err != nil {
		return bson.M{}, err
	}

	for cur.Next(ctx) {
		err := cur.Decode(&result)
		if err != nil {
			return bson.M{}, err
		}
	}

	if result != nil {
		return result["payload"].(primitive.M), nil
	} else {
		return bson.M{}, nil
	}
}

func (db *Database) Pool(faculty string) (map[string][]primitive.M, error) {
	var filter bson.M
	output := make(map[string][]primitive.M)
	db_connection := db.client.Database("rasp")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if faculty != "" {
		filter = bson.M{"name": faculty}
	} else {
		filter = bson.M{}
	}

	faculties, err := db_connection.ListCollectionNames(ctx, filter)
	if err != nil {
		return output, err
	}

	for _, faculty := range faculties {
		collection := db_connection.Collection(faculty)

		cur, err := collection.Find(ctx, bson.M{})
		if err != nil {
			return output, err
		}

		defer cur.Close(ctx)

		for cur.Next(ctx) {
			var result bson.M

			err := cur.Decode(&result)
			if err != nil {
				return output, err
			}

			delete(result, "_id")
			output[faculty] = append(output[faculty], result)
		}
	}

	return output, err
}

func (db *Database) FacultiesPool() ([]string, error) {
	db_connection := db.client.Database("rasp")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	faculties, err := db_connection.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return faculties, err
	}

	return faculties, err
}

func (db *Database) InsertToken(token string) (int, error) {
	collection := db.client.Database("db").Collection("tokens")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	itemCount, err := collection.CountDocuments(ctx, bson.M{"token": token})
	if err != nil {
		return -1, err
	}

	if itemCount >= 1 {
		return -1, nil
	} else {
		_, err := collection.InsertOne(ctx, bson.M{"token": token})
		if err != nil {
			return 0, err
		}

		return 1, nil
	}
}

func (db *Database) CheckToken(token string) (int, error) {
	collection := db.client.Database("db").Collection("tokens")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ItemCount, err := collection.CountDocuments(ctx, bson.M{"token": token})
	if err != nil {
		return -1, err
	}

	if ItemCount == 1 {
		return 1, nil
	} else {
		return -1, nil
	}
}

func db_init(address string, port string) (Database, error) {
	uri := "mongodb://" + address + ":" + port

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return Database{}, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return Database{}, err
	}

	return Database{client}, nil
}

func (db *Database) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.client.Disconnect(ctx); err != nil {
		return err
	}

	return nil
}
