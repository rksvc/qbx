package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func handleAPIVersion(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	json.NewEncoder(w).Encode(apiVer)
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	resp := make(map[LogType]struct {
		Session int64 `json:"session"`
		All     int64 `json:"all"`
	})
	mu.Lock()
	for typ, cnt := range stats {
		s := resp[typ]
		s.Session = cnt
		resp[typ] = s
	}
	mu.Unlock()

	for typ := BanInBlockedSubnet; typ <= BanSubnetTooManyPeers; typ++ {
		s := resp[typ]
		err := db.QueryRow(`select count(*) from logs where type = ?`, typ).Scan(&s.All)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp[typ] = s
	}
	json.NewEncoder(w).Encode(resp)
}

func handleLogs(w http.ResponseWriter, r *http.Request) {
	predicate := "1"
	if before := r.URL.Query().Get("before"); before != "" {
		predicate = "id < " + before
	}

	type Log struct {
		ID     int64     `json:"id"`
		Type   int       `json:"type"`
		Date   time.Time `json:"date"`
		Peer   string    `json:"peer"`
		Client string    `json:"client"`
	}
	logs := make([]Log, 0)

	rows, err := db.Query(fmt.Sprintf(`
		select id, type, date, peer, client
		from logs
		where %s
		order by id desc
		limit ?
	`, predicate), 15)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for rows.Next() {
		var lg Log
		err = rows.Scan(&lg.ID, &lg.Type, &lg.Date, &lg.Peer, &lg.Client)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logs = append(logs, lg)
	}
	if err = rows.Err(); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(logs)
}
