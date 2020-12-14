package main

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"context"
	"time"
)

const DatabaseTimeout = 20 * time.Second

type Database struct {
	client *mongo.Client
}

type DatabaseOptions struct {
	Address string
	Port    string
}

func (db *Database) Insert(FacultyName string, GroupName string, GroupRasp MongoSchedule) error {
	collection := db.client.Database(FacultyName).Collection(GroupName)

	ctx, cancel := context.WithTimeout(context.Background(), DatabaseTimeout)
	defer cancel()

	count, err := collection.CountDocuments(ctx, bson.M{"date": GroupRasp.Date})
	if err != nil {
		return err
	}

	if count == 0 {
		if _, err := collection.InsertOne(ctx, GroupRasp); err != nil {
			return err
		}
	} else {
		if _, err := collection.ReplaceOne(
			ctx,
			bson.M{"date": GroupRasp.Date},
			GroupRasp,
		); err != nil {
			return err
		}
	}

	return nil
}

func DbInit(cfg DatabaseOptions) (Database, error) {
	uri := "mongodb://" + cfg.Address + ":" + cfg.Port

	ctx, cancel := context.WithTimeout(context.Background(), DatabaseTimeout)
	defer cancel()

	db, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return Database{}, err
	}

	if err := db.Ping(ctx, readpref.Primary()); err != nil {
		return Database{}, err
	}

	return Database{db}, nil
}

func (db *Database) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), DatabaseTimeout)
	defer cancel()

	if err := db.client.Disconnect(ctx); err != nil {
		return err
	}

	return nil
}
