package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"text/template"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

const limit = 20

var rd = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

type D struct {
	Time []int64
	Data []string
}

func xlx(w http.ResponseWriter, r *http.Request) {

	var data D
	// The method must be get otherwise dump
	if r.Method != "GET" {
		w.WriteHeader(401)
		w.Write([]byte("Sorry, that method is not supported"))
		return
	}

	w.WriteHeader(200)

	keys, err := rd.Keys(ctx, "*").Result()

	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	if err != nil {
		fmt.Println(err)
	}

	for i := 0; i < len(keys); i++ {
		val, err := rd.Get(ctx, keys[i]).Result()

		if err != nil {
			fmt.Println(err)
		}

		time, _ := strconv.ParseInt(keys[i], 10, 64)

		data.Data = append(data.Data, val)

		data.Time = append(data.Time, time)

		if i == limit-1 {
			break
		}

	}

	fmt.Println(data)

	t, _ := template.ParseFiles("template/testpl.html")
	t.Execute(w, data)

}

func main() {

	go http.HandleFunc("/xlx", xlx)
	log.Fatal(http.ListenAndServe(":8080", nil))

}
