package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go.etcd.io/bbolt"
)

var (
	BUCKET_LOGS    = []byte("logs")
	BUCKET_STATS   = []byte("stats")
	BUCKET_SESSION = []byte("session")

	KEY_LAST_RESET     = []byte("last_reset")
	KEY_BANNED_SUBNETS = []byte("banned_subnets")
	KEY_BANNED_IPS     = []byte("banned_ips")
)

var db *bbolt.DB

func initDB(path string) (err error) {
	db, err = bbolt.Open(path, 0600, nil)
	if err != nil {
		return
	}
	return db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(BUCKET_LOGS)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(BUCKET_STATS)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(BUCKET_SESSION)
		return err
	})
}

func get(b *bbolt.Bucket, key []byte, value any) error {
	v := b.Get(key)
	if v == nil {
		return nil
	}
	return json.Unmarshal(v, value)
}

func put(b *bbolt.Bucket, key, value any) (err error) {
	var k []byte
	switch key := key.(type) {
	case []byte:
		k = key
	case time.Time:
		k = []byte(key.UTC().Format("2006-01-02T15:04:05.000000000Z07:00"))
	default:
		panic(fmt.Errorf("unsupported key type %T", key))
	}

	var v []byte
	switch value := value.(type) {
	case []byte:
		v = value
	default:
		v, err = json.Marshal(value)
		if err != nil {
			return
		}
	}

	return b.Put(k, v)
}

func inc(b *bbolt.Bucket, key LogType) error {
	count, err := load(b, key)
	if err != nil {
		return err
	}
	return b.Put(keyFromLogType(key), []byte(strconv.FormatInt(count+1, 10)))
}

func load(b *bbolt.Bucket, key LogType) (count int64, err error) {
	v := b.Get(keyFromLogType(key))
	if v != nil {
		count, err = strconv.ParseInt(string(v), 10, 64)
	}
	return
}

func del(b *bbolt.Bucket, key LogType) error {
	return b.Delete(keyFromLogType(key))
}

func keyFromLogType(key LogType) []byte {
	return []byte(strconv.Itoa(int(key)))
}
