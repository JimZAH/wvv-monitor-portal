package main

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

var rd = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

func main() {
	data, err := rd.Keys(ctx, "*").Result()

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(data)

}
