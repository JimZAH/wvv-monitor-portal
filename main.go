package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

var rd = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

func testpl(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	t, _ := template.ParseFiles("template/testpl.html")
	t.Execute(w, nil)
}

func xlx(w http.ResponseWriter, r *http.Request) {

	// The method must be get otherwise dump
	if r.Method != "GET" {
		w.WriteHeader(401)
		w.Write([]byte("Sorry, that method is not supported"))
		return
	}

	w.WriteHeader(200)

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
		w.Write([]byte(data[i] + "\n-----------------------------------------------------------------------\n"))
	}

}

func main() {

	http.HandleFunc("/xlx", xlx)
	http.HandleFunc("/test", testpl)
	log.Fatal(http.ListenAndServe(":8080", nil))

}
