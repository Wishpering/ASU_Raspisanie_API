package main

import (
       "go.mongodb.org/mongo-driver/mongo"
       "go.mongodb.org/mongo-driver/bson/primitive"
)

type Database struct {
     client *mongo.Client
}

type Token struct {
     Key string
}

type Pool struct {
     Count int
     Payload map[string] primitive.M
}

type Configuration struct {
     Password string
     Port string
     Compress bool
     DB_address string
     DB_port string
}