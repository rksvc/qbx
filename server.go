package main

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v5"
	"go.etcd.io/bbolt"
)

func handleAPIVersion(c *echo.Context) error {
	mu.Lock()
	defer mu.Unlock()
	return c.JSON(http.StatusOK, apiVer)
}

func handleStats(c *echo.Context) error {
	resp := make(map[LogType]struct {
		Session int64 `json:"session"`
		All     int64 `json:"all"`
	})

	err := db.View(func(tx *bbolt.Tx) error {
		stats := tx.Bucket(BUCKET_STATS)
		session := tx.Bucket(BUCKET_SESSION)
		for key := range banReasons() {
			var err error
			s := resp[key]
			if s.All, err = load(stats, key); err != nil {
				return err
			}
			if s.Session, err = load(session, key); err != nil {
				return err
			}
			resp[key] = s
		}
		return nil
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, resp)
}

func handleLogs(c *echo.Context) error {
	type Log struct {
		Date       string `json:"date"`
		Type       int    `json:"t"`
		Peer       string `json:"p,omitempty"`
		Client     string `json:"c,omitempty"`
		Downloaded int64  `json:"d,omitempty"`
		Uploaded   int64  `json:"u,omitempty"`
	}
	logs := make([]Log, 0)

	err := db.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(BUCKET_LOGS).Cursor()
		var key, value []byte
		if before := c.QueryParam("before"); before == "" {
			key, value = cursor.Last()
		} else {
			key, _ = cursor.Seek([]byte(before))
			if key != nil {
				key, value = cursor.Prev()
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

			key, value = cursor.Prev()
		}
		return nil
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, logs)
}
