package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/go-redis/redis/v8"
)

var (
	ctx   = context.Background()
	limit = 50
)

var rd = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

type D struct {
	Time   []int64
	Data   []string
	Reload bool
}

func xlx(w http.ResponseWriter, r *http.Request) {

	var data D
	data.Reload = true

	// The method must be get otherwise dump
	if r.Method != "GET" {
		w.WriteHeader(401)
		w.Write([]byte("Sorry, that method is not supported"))
		return
	}

	span := r.URL.Query().Get("xlx_span")

	if span != "" {
		data.Reload = false
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

		// Remove when migration is complete. If the data is raw skip!
		raw := strings.Split(keys[i], "-")
		if raw[0] == "raw" {
			continue
		}

		time, _ := strconv.ParseInt(keys[i], 10, 64)

		data.Data = append(data.Data, val)

		data.Time = append(data.Time, time)

		if i == limit-1 && span != "full" {
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
