package main

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

// 定义一个全局的pool
var pool *redis.Pool

func initPool(address string, maxIdle int, maxActive int, idleTimeout time.Duration) {
	pool = &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: idleTimeout,
		Dial: func() (redis.Conn, error) { //初始化连接
			return redis.Dial("tcp", address)
		},
	}
}
