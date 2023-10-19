package main

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Config struct {
	Hosts             string
	Name              string
	Username          string
	Password          string
	AuthMechanism     string
	ReplicaSet        string
	AdminName         string
	ReadTimeout       time.Duration
	ConnectionTimeout time.Duration
	MaxConnIdleTime   time.Duration
	PoolLimit         int
	ReadSecondaryPref bool
}

func (cfg Config) FinalizeClientOptions() *options.ClientOptions {
	opts := options.Client().
		SetHosts(strings.Split(cfg.Hosts, ",")).
		SetConnectTimeout(cfg.ConnectionTimeout).
		SetSocketTimeout(cfg.ReadTimeout).
		SetMaxConnIdleTime(cfg.MaxConnIdleTime)

	if cfg.Username != "" || cfg.Password != "" {
		opts.SetAuth(options.Credential{
			AuthSource:    cfg.AdminName,
			AuthMechanism: cfg.AuthMechanism,
			Username:      cfg.Username,
			Password:      cfg.Password,
		})
	}

	if cfg.ReplicaSet != "" {
		opts.SetReplicaSet(cfg.ReplicaSet)
	}

	if cfg.PoolLimit > 0 {
		opts.SetMaxPoolSize(uint64(cfg.PoolLimit))
	}

	if cfg.ReadSecondaryPref {
		opts.SetReadPreference(readpref.SecondaryPreferred())
	}

	return opts
}

func NewMongoClient(ctx context.Context, cfg Config) (*mongo.Client, error) {
	opts := cfg.FinalizeClientOptions()

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	return client, nil

}
