package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

var rd = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

type D struct {
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

	if err != nil {
		fmt.Println(err)
	}

	for i := 0; i < len(keys); i++ {
		val, err := rd.Get(ctx, keys[i]).Result()

		if err != nil {
			fmt.Println(err)
		}

		data.Data = append(data.Data, val)

	}

	fmt.Println(data)

	t, _ := template.ParseFiles("template/testpl.html")
	t.Execute(w, data)

}

func main() {

	http.HandleFunc("/xlx", xlx)
	log.Fatal(http.ListenAndServe(":8080", nil))

}
