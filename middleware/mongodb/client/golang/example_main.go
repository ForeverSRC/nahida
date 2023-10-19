package main

import (
	"context"
	"time"
)

type Student struct {
	StudentID int         `bson:"student_id"`
	ClassID   int         `bson:"class_id"`
	Scores    []ScoreItem `bson:"scores"`
}

type ScoreItem struct {
	Type  string  `bson:"type"`
	Score float64 `bson:"score"`
}

func main() {
	cfg := Config{
		Hosts:             "localhost:27017",
		Username:          "root",
		Password:          "aaa1234",
		ReadTimeout:       30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
		MaxConnIdleTime:   30 * time.Second,
	}

	client, err := NewMongoClient(context.Background(), cfg)
	if err != nil {
		panic(err)
	}

	db := client.Database("test")

	collection := db.Collection("students")

	for i := 0; i < 100000; i++ {
		s := Student{
			StudentID: i + 1,
			ClassID:   (i + 1) % 100,
			Scores: []ScoreItem{
				{
					Type:  "exam",
					Score: float64((i + 50) % 100),
				},
				{
					Type:  "quiz",
					Score: float64(i % 10),
				},
			},
		}

		_, err := collection.InsertOne(context.Background(), s)
		if err != nil {
			panic(err)
		}

	}
}
