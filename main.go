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
	var data []string

	keys, err := rd.Keys(ctx, "*").Result()

	if err != nil {
		fmt.Println(err)
	}

	for i := 0; i < len(keys); i++ {
		val, err := rd.Get(ctx, keys[i]).Result()

		if err != nil {
			fmt.Println(err)
		}

		data = append(data, val)

	}

	for i := 0; i < len(data); i++ {
		fmt.Println(data[i])
	}

}
