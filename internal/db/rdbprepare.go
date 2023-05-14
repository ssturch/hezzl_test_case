package db

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisClient struct {
	ctx       context.Context
	rdsClient *redis.Client
}

const (
	hostRds     = "localhost"
	portRds     = "6379"
	passwordRds = ""
	dbRds       = 0
)

//func (r *RedisClient) Prepare(ctx context.Context) {
//	r.ctx = ctx
//}

func (r *RedisClient) Invalidate() {
	r.rdsClient.Del(r.ctx, "list")
}

func (r *RedisClient) AddData(val []byte) {
	r.rdsClient.Set(r.ctx, "list", val, time.Minute)
}

func (r *RedisClient) GetData() ([]byte, error) {
	res, err := r.rdsClient.Get(r.ctx, "list").Bytes()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func PrepareRedis(ctx context.Context) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", hostRds, portRds),
		Password: passwordRds,
		DB:       dbRds,
	})
	m := &RedisClient{rdsClient: client, ctx: ctx}
	return m
}
