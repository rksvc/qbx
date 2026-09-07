package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"go.etcd.io/bbolt"
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

	err := db.View(func(tx *bbolt.Tx) error {
		getStats := func(bucket string, session bool) error {
			stats := tx.Bucket([]byte(bucket))
			for t := range logTypes() {
				value := stats.Get([]byte(strconv.Itoa(int(t))))
				if value != nil {
					s := resp[t]
					var err error
					if session {
						s.Session, err = strconv.ParseInt(string(value), 10, 64)
					} else {
						s.All, err = strconv.ParseInt(string(value), 10, 64)
					}
					if err != nil {
						return err
					}
					resp[t] = s
				}
			}
			return nil
		}
		err := getStats(BUCKET_STATS, false)
		if err != nil {
			return err
		}
		return getStats(BUCKET_SESSION, true)
	})
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(resp)
}

func handleLogs(w http.ResponseWriter, r *http.Request) {
	type Log struct {
		Date   string `json:"d"`
		Type   int    `json:"t"`
		Peer   string `json:"p,omitempty"`
		Client string `json:"c,omitempty"`
	}
	logs := make([]Log, 0)

	err := db.View(func(tx *bbolt.Tx) error {
		c := tx.Bucket([]byte("logs")).Cursor()
		var key, value []byte
		if before := r.URL.Query().Get("before"); before == "" {
			key, value = c.Last()
		} else {
			key, _ = c.Seek([]byte(before))
			if key != nil {
				key, value = c.Prev()
			}
		}

		for range 15 {
			if key == nil {
				break
			}

			var log Log
			err := json.Unmarshal(value, &log)
			if err != nil {
				return err
			}
			log.Date = string(key)
			logs = append(logs, log)

			key, value = c.Prev()
		}
		return nil
	})
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(logs)
}
