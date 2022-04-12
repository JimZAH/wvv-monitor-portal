package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	ctx   = context.Background()
	limit = 50
	ip    []Ip
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

type Ip struct {
	Count     int64
	IPAddress string
	LTime     int64
	STime     int64
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
	Epoch         int64  `json:"Epoch"`
}

func limiter(ipaddress string) bool {

	for i := 0; i < len(ip); i++ {

		if time.Now().Unix()-ip[i].STime > 60 {
			log.Println("FLUSH Count: ", ip[i].IPAddress)
			ip[i].STime = time.Now().Unix()
			ip[i].Count = 0
		}

		if ipaddress == ip[i].IPAddress {
			ip[i].Count++
			ip[i].LTime = time.Now().Unix()
			if ip[i].Count > 10*3 {
				log.Println("IP Limit: ", ip[i].IPAddress, ip[i].Count)
				return false
			}
			return true
		}
	}

	nip := Ip{0, ipaddress, time.Now().Unix(), time.Now().Unix()}
	ip = append(ip, nip)
	log.Println("Found:", ipaddress)

	return true
}

func xlx(w http.ResponseWriter, r *http.Request) {

	var data D
	data.Reload = true

	// The method must be get otherwise dump
	if r.Method != "GET" {
		w.WriteHeader(405)
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
		log.Println(err)
	}

	for i := 0; i < len(keys); i++ {
		val, err := rd.Get(ctx, keys[i]).Result()

		if err != nil {
			log.Println(err)
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

	log.Println(data)

	t, _ := template.ParseFiles("template/testpl.html")
	t.Execute(w, data)

}

func xlxJson(w http.ResponseWriter, r *http.Request) {

	if !limiter(r.Header.Get("Forwarded")) {
		w.WriteHeader(429)
		return
	}

	if r.Method != "GET" {
		w.WriteHeader(405)
	}

	var (
		c   int64 = 0
		d   Station
		max int64
		s   []Station
	)

	limit := false

	w.Header().Set("Content-Type", "application/json")

	maxParam := r.URL.Query().Get("max")

	if maxParam != "" {
		var err error
		limit = true
		max, err = strconv.ParseInt(maxParam, 10, 64)
		if err != nil {
			log.Println(err)
		}

		if max < 1 {
			w.WriteHeader(416)
			return
		}
	}

	w.WriteHeader(200)

	keys, err := rd.Keys(ctx, "*").Result()
	if err != nil {
		log.Println(err)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	for i := 0; i < len(keys); i++ {
		val, err := rd.Get(ctx, keys[i]).Result()
		if err != nil {
			log.Println(err)
		}

		// This won't be required on final release
		spl := strings.Split(keys[i], "-")
		if spl[0] != "raw" {
			continue
		}

		json.Unmarshal([]byte(val), &d)

		time, _ := strconv.ParseInt(spl[1], 10, 64)

		NewStation := Station{
			d.Callsign,
			d.Vianode,
			d.Onmodule,
			d.Viapeer,
			d.LastHeardTime,
			time}

		s = append(s, NewStation)

		c++

		// if limit is reached
		if c >= max && limit {
			break
		} else if c >= 50 && !limit {
			break
		}
	}
	json.NewEncoder(w).Encode(s)
}

func xlxNodesJson(w http.ResponseWriter, r *http.Request) {

	var node []Node

	if !limiter(r.Header.Get("Forwarded")) {
		w.WriteHeader(429)
		return
	}

	if r.Method != "GET" {
		w.WriteHeader(405)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	nodes, err := rd.Get(ctx, "nodes").Result()
	if err != nil {
		log.Println(err)
	}

	json.Unmarshal([]byte(nodes), &node)

	json.NewEncoder(w).Encode(node)

}

func main() {

	go http.HandleFunc("/xlx", xlx)
	go http.HandleFunc("/xlx-stations", xlxJson)
	go http.HandleFunc("/xlx-nodes", xlxNodesJson)
	log.Fatal(http.ListenAndServe(":8080", nil))

}
