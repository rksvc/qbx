package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/Masterminds/semver"
)

func ok(resp *http.Response) bool {
	if resp.StatusCode < 400 {
		return true
	}
	log.Printf("%s %s: %s", resp.Request.Method, resp.Request.URL, resp.Status)
	return false
}

func login() bool {
	resp, err := client.PostForm(addr+"/api/v2/auth/login", url.Values{
		"username": {username},
		"password": {password},
	})
	if err != nil {
		log.Print(err)
		return false
	}
	defer resp.Body.Close()
	return ok(resp)
}

func checkAPIVersion() (bool, int) {
	resetAPIVer := func() {
		mu.Lock()
		apiVer.Version = ""
		apiVer.Supported = false
		mu.Unlock()
	}

	resp, err := client.Get(addr + "/api/v2/app/webapiVersion")
	if err != nil {
		log.Print(err)
		resetAPIVer()
		return false, 0
	}
	defer resp.Body.Close()
	if !ok(resp) {
		resetAPIVer()
		return false, resp.StatusCode
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
		resetAPIVer()
		return false, http.StatusOK
	}
	ver := string(b)

	mu.Lock()
	defer mu.Unlock()
	apiVer.Version = ver
	apiVer.Supported = !semver.MustParse(ver).LessThan(semver.MustParse("2.3"))
	if !apiVer.Supported {
		log.Print("qBittorrent API >= v2.3 required")
		return false, http.StatusOK
	}
	return true, http.StatusOK
}

func clearBannedIPs() {
	resp, err := client.PostForm(addr+"/api/v2/app/setPreferences", url.Values{
		"json": {`{"banned_IPs":""}`},
	})
	if err != nil {
		log.Print(err)
		return
	}
	defer resp.Body.Close()
	ok(resp)
}

func get(url string, value any) bool {
	resp, err := client.Get(url)
	if err != nil {
		log.Print(err)
		return false
	}
	defer resp.Body.Close()
	if !ok(resp) {
		return false
	}

	err = json.NewDecoder(resp.Body).Decode(value)
	if err != nil {
		log.Print(err)
		return false
	}
	return true
}

func banPeers(peers string) {
	resp, err := client.PostForm(addr+"/api/v2/transfer/banPeers", url.Values{
		"peers": {peers},
	})
	if err != nil {
		log.Print(err)
		return
	}
	defer resp.Body.Close()
	ok(resp)
}
