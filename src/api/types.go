package main

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Database struct {
	client *mongo.Client
}

type Token struct {
	Key []byte `json:"key"`
}

type Pool struct {
	Count   int `json:"count"`
	Payload map[string]primitive.M `json:"payload"`
}

type Configuration struct {
	Password   []byte
	Port       string
	Compress   bool
	DB_address string
	DB_port    string
}
