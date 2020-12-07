package main

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"context"
	"time"
)

const ContextTimeout = 15 * time.Second

type Database struct {
	Client *mongo.Client
}

func (db *Database) GetScheduleByDate(faculty string, group_id string, date string) (bson.M, error) {
	var (
		record bson.M
		filter bson.M
	)

	collection := db.Client.Database(faculty).Collection(group_id)

	ctx, cancel := context.WithTimeout(context.Background(), ContextTimeout)
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

func (db *Database) Pool() ([]Faculty, error) {
	var buffer []Faculty

	ctx, cancel := context.WithTimeout(context.Background(), ContextTimeout)
	defer cancel()

	if Faculties, err := db.Client.ListDatabaseNames(ctx, bson.M{}); err != nil {
		return buffer, err
	} else {
		for _, faculty := range Faculties {
			if faculty == "admin" || faculty == "local" || faculty == "config" {
				continue
			}

			db := db.Client.Database(faculty)

			if groups, err := db.ListCollectionNames(ctx, bson.M{}); err != nil {
				return buffer, err
			} else {
				groupsPoolStruct := GroupsPool{len(groups), groups}
				
				buffer = append(
					buffer,
					Faculty{
						faculty,
						groupsPoolStruct,
					},
				)
			}
		}
	}

	return buffer, nil
}

func (db *Database) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), ContextTimeout)
	defer cancel()

	if err := db.Client.Disconnect(ctx); err != nil {
		return err
	}

	return nil
}

func InitDB(address string, port string) (Database, error) {
	var db Database
	
	uri := "mongodb://" + address + ":" + port

	ctx, cancel := context.WithTimeout(context.Background(), ContextTimeout)
	defer cancel()

	if client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri)); err != nil {
		return db, err
	} else {
		if err := client.Ping(ctx, readpref.Primary()); err != nil {
			return db, err
		}
		
		db.Client = client
	}
	
	return db, nil
}
