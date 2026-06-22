package main

import (
	"database/sql"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/publicsuffix"
)

type LogType int

const (
	ClearBannedIPs LogType = iota
	BanInBlockedSubnet
	BanWeirdClient
	BanLeecherClient
	BanObsoleteClient
	BanUploadedMoreThanTotalSize
	BanNoProgress
	BanShrunkProgress
	BanUploadedExcessively
	BanSubnetTooManyPeersBanned
	BanSubnetTooManyPeers
)

//go:embed dist
var dist embed.FS

var (
	leecherClients  = []string{"-XL", "Xunlei", "XunLei", "7.", "aria2", "Xfplay", "dandanplay", "FDM", "go.torrent", "Mozilla", "github.com/anacrolix/torrent (devel) (anacrolix/torrent unknown)", "dt/torrent/", "Taipei-Torrent dev", "trafficConsume", "hp/torrent/", "BitComet 1.92", "BitComet 1.98", "xm/torrent/", "flashget", "FlashGet", "StellarPlayer", "Gopeed", "MediaGet", "aD/", "ADM", "coc_coc_browser", "FileCroc", "filecxx", "Folx", "seanime (devel) (anacrolix/torrent", "HitomiDownloader", "gateway (devel) (anacrolix/torrent", "offline-download", "QQDownload", "git.woa.com", "iLivid"}
	obsoleteClients = []string{"TorrentStorm", "Azureus 1.", "Azureus 2.", "Azureus 3.", "Deluge 0.", "Deluge 1.0", "Deluge 1.1", "qBittorrent 0.", "qBittorrent 1.", "qBittorrent 2.", "Transmission 0.", "Transmission 1.", "BitComet 0.", "µTorrent 1.", "uTorrent 1.", "μTorrent 1."}
)

var (
	conf, addr, username, password, webui string

	mu     sync.Mutex
	apiVer struct {
		Version   string `json:"version"`
		Supported bool   `json:"supported"`
	}
	stats = make(map[LogType]int64)

	db     *sql.DB
	client *http.Client
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.StringVar(&conf, "conf", "", "config directory")
	flag.StringVar(&addr, "addr", "http://127.0.0.1:8080", "qBittorrent WebUI address")
	flag.StringVar(&username, "username", "", "qBittorrent WebUI username")
	flag.StringVar(&password, "password", "", "qBittorrent WebUI password")
	flag.StringVar(&webui, "webui", ":2386", "WebUI address")
	flag.Parse()

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}
	client = &http.Client{Jar: jar}

	if conf == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			log.Fatal(err)
		}
		conf = filepath.Join(dir, "qbx")
	}
	if err = os.MkdirAll(conf, 0755); err != nil {
		log.Fatal(err)
	}
	db, err = sql.Open("sqlite3", path.Join(conf, "qbx.db"))
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`
		create table if not exists logs (
			id integer primary key autoincrement,
			type integer,
			date datetime,
			peer text,
			client text
		);
	`)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		dist, _ := fs.Sub(dist, "dist")
		http.HandleFunc("/api/apiVersion", handleAPIVersion)
		http.HandleFunc("/api/stats", handleStats)
		http.HandleFunc("/api/logs", handleLogs)
		http.Handle("/", http.FileServer(http.FS(dist)))
		log.Fatal(http.ListenAndServe(webui, nil))
	}()
}

func xlog(typ LogType, peer, client string) {
	_, err := db.Exec(`
		insert into logs (type, date, peer, client)
		values (?, ?, ?, ?)
	`, typ, time.Now().UTC(), peer, client)
	if err != nil {
		log.Fatal(err)
	}
	mu.Lock()
	stats[typ]++
	mu.Unlock()
}

func main() {
	type (
		Subnet = string
		Host   = string
	)
	var (
		torrents = make(map[string]struct {
			Private              bool
			TotalSize, PieceSize int64
			PeerInitialProgress  map[string]float64
		})
		bannedSubnets = make(map[Subnet]struct{})
		bannedIPs     = make(map[Subnet]map[Host]struct{})
	)

	var lastReset time.Time
	for ticker := time.NewTicker(10 * time.Second); true; <-ticker.C {
		ok, statusCode := checkAPIVersion()
		if statusCode == http.StatusForbidden && login() {
			ok, statusCode = checkAPIVersion()
		}
		if !ok {
			continue
		}

		if time.Since(lastReset).Hours() >= 24 {
			lastReset = time.Now()
			xlog(ClearBannedIPs, "", "")
			clear(torrents)
			clear(bannedSubnets)
			clear(bannedIPs)
			clear(stats)
			clearBannedIPs()
		}

		var mainData struct {
			Torrents map[string]struct {
				TotalSize int64 `json:"total_size"`
			} `json:"torrents"`
		}
		if !get(addr+"/api/v2/sync/maindata", &mainData) {
			continue
		}
		var bannedPeers []string
		for hash, torrent := range mainData.Torrents {
			if torrent.TotalSize < 0 {
				continue
			}

			t, ok := torrents[hash]
			if !ok {
				var trackers []struct {
					Status int `json:"status"`
				}
				if !get(addr+"/api/v2/torrents/trackers?hash="+hash, &trackers) {
					continue
				}
				// DHT is disabled and number of trackers <= 3
				t.Private = trackers[0].Status == 0 && len(trackers) >= 4 && len(trackers) <= 6
				if !t.Private {
					var properties struct {
						PieceSize int64 `json:"piece_size"`
					}
					if !get(addr+"/api/v2/torrents/properties?hash="+hash, &properties) {
						continue
					}
					t.PieceSize = properties.PieceSize
					t.TotalSize = torrent.TotalSize
					t.PeerInitialProgress = make(map[string]float64)
				}
				torrents[hash] = t
			}
			if t.Private {
				continue
			}

			var (
				peerSubnets  = make(map[Subnet]map[Host]struct{})
				peerToClient = make(map[string]string)
			)

			var peers struct {
				Peers map[string]struct {
					Client     string  `json:"client"`
					Downloaded int64   `json:"downloaded"`
					Flags      string  `json:"flags"`
					Progress   float64 `json:"progress"`
					Uploaded   int64   `json:"uploaded"`
				} `json:"peers"`
			}
			if !get(addr+"/api/v2/sync/torrentPeers?hash="+hash, &peers) {
				continue
			}
			for ip, peer := range peers.Peers {
				var banPeer bool
				ban := func(reason LogType) {
					xlog(reason, ip, peer.Client)
					banPeer = true
				}

				subnet := ip
				sep := byte('.')
				if after, found := strings.CutPrefix(subnet, "[::ffff:"); found {
					subnet = after
				} else if subnet[0] == '[' {
					sep = ':'
				}
				var idx int
				for range 3 {
					idx += strings.IndexByte(subnet[idx:], sep) + 1
				}
				subnet = subnet[:idx]

				if _, ok := bannedSubnets[subnet]; ok {
					ban(BanInBlockedSubnet)
				} else {
					hosts, ok := peerSubnets[subnet]
					if !ok {
						hosts = make(map[string]struct{})
						peerSubnets[subnet] = hosts
					}
					hosts[ip] = struct{}{}
					peerToClient[ip] = peer.Client
				}

				if !banPeer {
					for _, c := range peer.Client {
						if c < ' ' || (c > '~' && c != 'µ' && c != 'μ') {
							ban(BanWeirdClient)
							break
						}
					}
				}

				if !banPeer && (peer.Progress == 0 || (peer.Downloaded == 0 && peer.Progress != 1)) && peer.Client != "" {
					if len(peer.Client) < 4 || peer.Client[2] == ' ' || strings.HasPrefix(peer.Client, "Unknown") {
						ban(BanWeirdClient)
					} else {
						for _, c := range leecherClients {
							if strings.HasPrefix(peer.Client, c) {
								ban(BanLeecherClient)
								break
							}
						}
						if !banPeer {
							for _, c := range obsoleteClients {
								if strings.HasPrefix(peer.Client, c) {
									ban(BanObsoleteClient)
									break
								}
							}
						}
					}
				}

				if !banPeer && strings.ContainsRune(peer.Flags, 'U') {
					if peer.Uploaded > t.TotalSize+2*t.PieceSize {
						ban(BanUploadedMoreThanTotalSize)
					} else if peer.Progress == 0 && peer.Uploaded > 10*1024*1024 && peer.Uploaded > 2*t.PieceSize {
						ban(BanNoProgress)
					} else if initialProgress, ok := t.PeerInitialProgress[ip]; !ok {
						t.PeerInitialProgress[ip] = max(0, peer.Progress-float64(peer.Uploaded)/float64(t.TotalSize))
					} else if peer.Progress < initialProgress {
						ban(BanShrunkProgress)
					} else if boundarySize := int64(30 * 1024 * 1024); t.TotalSize > boundarySize &&
						peer.Uploaded-boundarySize > int64(float64(t.TotalSize)*(peer.Progress-initialProgress)) {
						ban(BanUploadedExcessively)
					}
				}

				if banPeer {
					bannedPeers = append(bannedPeers, ip)
					if _, ok := bannedSubnets[subnet]; !ok {
						ips, ok := bannedIPs[subnet]
						if !ok {
							ips = make(map[string]struct{})
							bannedIPs[subnet] = ips
						}
						ips[ip] = struct{}{}

						if len(ips) == 5 {
							var suffix string
							if subnet[0] == '[' {
								suffix = "]"
							}
							xlog(BanSubnetTooManyPeersBanned, fmt.Sprintf("%s*%s", subnet, suffix), "")
							bannedSubnets[subnet] = struct{}{}
							delete(bannedIPs, subnet)
							delete(peerSubnets, subnet)
						}
					}
				}
			}

			for subnet, ips := range peerSubnets {
				if len(ips) >= 5 {
					var suffix string
					if subnet[0] == '[' {
						suffix = "]"
					}
					xlog(BanSubnetTooManyPeers, fmt.Sprintf("%s*%s", subnet, suffix), "")
					bannedSubnets[subnet] = struct{}{}
					for ip := range ips {
						xlog(BanInBlockedSubnet, ip, peerToClient[ip])
						bannedPeers = append(bannedPeers, ip)
					}
				}
			}
		}

		if len(bannedPeers) > 0 {
			banPeers(strings.Join(bannedPeers, "|"))
		}
	}
}
