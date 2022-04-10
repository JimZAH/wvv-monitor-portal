package main

import (
	"context"
	"encoding/json"
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

type Data struct {
	Key     string   `json:"Key"`
	Node    Nodes    `json:"Node"`
	Station Stations `json:"Station"`
}

type Nodes struct {
	N []Node `json:"Nodes"`
}

type Stations struct {
	S []Station `json:"Stations"`
}

type Node struct {
	Callsign      string `json:"Callsign"`
	IP            string `json:"IP"`
	LinkedModule  string `json:"LinkedModule"`
	Protocol      string `json:"Protocol"`
	ConnectTime   string `json:"ConnectTime"`
	LastHeardTime string `json:"LastHeardTime"`
}

type Station struct {
	Callsign      string `json:"Callsign"`
	Vianode       string `json:"Via-node"`
	Onmodule      string `json:"On-module"`
	Viapeer       string `json:"Via-peer"`
	LastHeardTime string `json:"LastHeardTime"`
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

func xlxJson(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(401)
	}

	var d Data

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	keys, err := rd.Keys(ctx, "*raw").Result()
	if err != nil {
		fmt.Println(err)
	}

	for i := 0; i < len(keys); i++ {
		val, err := rd.Get(ctx, keys[i]).Result()
		if err != nil {
			fmt.Println(err)
		}

		spl := strings.Split(keys[i], "-")
		if spl[0] != "raw" {
			continue
		}

		json.Unmarshal([]byte(val), &d)
	}
	json.NewEncoder(w).Encode(d)
}

func main() {

	go http.HandleFunc("/xlx", xlx)
	go http.HandleFunc("/org/xlx", xlxJson)
	log.Fatal(http.ListenAndServe(":8080", nil))

}
