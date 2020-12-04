package main

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"context"
	"time"
)

type Database struct {
	client *mongo.Client
}

func (db *Database) GetScheduleByDate(faculty string, group_id string, date string) (bson.M, error) {
	var (
		record bson.M
		filter bson.M
	)

	collection := db.client.Database(faculty).Collection(group_id)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if date != "" {
		filter = bson.M{"date": date}
	}

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return record, err
	}

	for cur.Next(ctx) {
		if err := cur.Decode(&record); err != nil {
			return record, err
		} else {
			delete(record, "_id")
		}
	}

	return record, nil
}

func (db *Database) Pool() (map[string][]string, error) {
	buffer := make(map[string][]string)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	Faculties, err := db.client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		return buffer, err
	}

	for _, faculty := range Faculties {
		if faculty == "admin" || faculty == "local" || faculty == "config" || faculty == "colly" {
			continue
		}

		db := db.client.Database(faculty)

		if groups, err := db.ListCollectionNames(ctx, bson.M{}); err != nil {
			return buffer, err
		} else {
			buffer[faculty] = groups
		}
	}

	return buffer, err
}

func (db *Database) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.client.Disconnect(ctx); err != nil {
		return err
	}

	return nil
}

func InitDB(address string, port string) (Database, error) {
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
