package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
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
	causalConsistency()
}

func simpleInsert() {
	cfg := Config{
		Hosts:             "mongo-rs0:27017,mongo-rs1:27018,mongo-rs2:27019",
		ReplicaSet:        "mdbDefGuide",
		Username:          "",
		Password:          "",
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

type Item struct {
	Name  string     `bson:"name"`
	SKU   string     `bson:"sku"`
	Start *time.Time `bson:"start"`
	End   *time.Time `bson:"end"`
}

func causalConsistency() {
	ctx := context.Background()
	cfg := Config{
		Hosts:             "mongo-rs0:27017,mongo-rs1:27018,mongo-rs2:27019",
		ReplicaSet:        "mdbDefGuide",
		Username:          "",
		Password:          "",
		ReadTimeout:       30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
		MaxConnIdleTime:   30 * time.Second,
	}

	client, err := NewMongoClient(ctx, cfg)
	if err != nil {
		panic(err)
	}

	coll := client.Database("test").Collection("items")

	currentDate := time.Now()

	// Start Causal Consistency Example 1

	// Use a causally-consistent session to run some operations
	opts := options.Session().SetDefaultReadConcern(readconcern.Majority()).SetDefaultWriteConcern(
		writeconcern.New(writeconcern.WMajority(), writeconcern.WTimeout(1000)))
	session1, err := client.StartSession(opts)
	if err != nil {
		panic(err)
	}
	defer session1.EndSession(context.TODO())

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		err = mongo.WithSession(ctx, session1, func(sctx mongo.SessionContext) error {
			// Run an update with our causally-consistent session
			_, err = coll.UpdateOne(sctx, bson.D{{"sku", "111"}}, bson.D{{"$set", bson.D{{"end", currentDate}}}})
			if err != nil {
				return err
			}
			fmt.Printf("session 1 updateOne cluster time: %v\n", session1.ClusterTime().String())
			fmt.Printf("session 1 updateOne operation time: %+v\n", session1.OperationTime())

			time.Sleep(1 * time.Millisecond)
			// Run an insert with our causally-consistent session
			_, err = coll.InsertOne(sctx, bson.D{{"sku", "nuts-111"}, {"name", "Pecans"}, {"start", currentDate}})
			if err != nil {
				return err
			}
			fmt.Printf("session 1 insertOne cluster time: %v\n", session1.ClusterTime().String())
			fmt.Printf("session 1 insertOne operation time: %+v\n", session1.OperationTime())

			return nil
		})

		if err != nil {
			panic(err)
		}
	}()

	// End Causal Consistency Example 1

	// Start Causal Consistency Example 2

	// Make a new session that is causally consistent with session1 so session2 reads what session1 writes
	opts = options.Session().SetDefaultReadPreference(readpref.Secondary()).SetDefaultReadConcern(
		readconcern.Majority()).SetDefaultWriteConcern(writeconcern.New(writeconcern.WMajority(),
		writeconcern.WTimeout(1000)))
	session2, err := client.StartSession(opts)
	if err != nil {
		panic(err)
	}
	defer session2.EndSession(ctx)

	go func() {
		defer wg.Done()
		err = mongo.WithSession(ctx, session2, func(sctx mongo.SessionContext) error {
			time.Sleep(5 * time.Millisecond)
			// Set cluster time of session2 to session1's cluster time
			clusterTime := session1.ClusterTime()
			session2.AdvanceClusterTime(clusterTime)
			fmt.Printf("session 2 advance cluster time: %v\n", clusterTime.String())

			// Set operation time of session2 to session1's operation time
			operationTime := session1.OperationTime()
			session2.AdvanceOperationTime(operationTime)
			fmt.Printf("session 2 advance opreation time: %+v\n", operationTime)
			// Run a find on session2, which should find all the writes from session1
			time.Sleep(5 * time.Millisecond)
			cursor, err := coll.Find(sctx, bson.D{{"end", nil}})

			if err != nil {
				return err
			}

			for cursor.Next(sctx) {
				doc := cursor.Current
				fmt.Printf("Document: %v\n", doc.String())
				fmt.Printf("session 2 cursor cluster time: %v\n", session2.ClusterTime().String())
				fmt.Printf("session 2 cursor operation time: %+v\n", session2.OperationTime())
			}

			return cursor.Err()
		})

		if err != nil {
			panic(err)
		}
		// End Causal Consistency Example 2
	}()

	wg.Wait()

}
