package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/cceckman/reading-list/server"
	"github.com/cceckman/reading-list/server/entry"
	serverfs "github.com/cceckman/reading-list/server/fs"
	serverLog "github.com/cceckman/reading-list/server/log"
	"github.com/cceckman/reading-list/server/paths"

	"tailscale.com/client/tailscale"
	"tailscale.com/tsnet"
)

var (
	addr       = flag.String("addr", ":443", "Port or address:port to listen on")
	allowLocal = flag.Bool("allowLocal", false, "Allow serving from the local static/ directory rather than embedded content. Development only.")
	storageDir = flag.String("storageDir", "", "Directory to use for entry management. If empty, uses an in-memory entry store.")
	useTsNet   = flag.Bool("tsnet", true, "Connect directly to Tailscale via tsnet")
)

func getEntryManager() server.EntryManager {
	if *storageDir == "" {
		m := &entry.TestEntryManager{
			Items: make(map[string]*entry.Entry),
		}
		m.Items["dmenu-menus"] = &entry.Entry{
			Id:    "dmenu-menus",
			Title: "Using dmenu to Optimize Common Tasks",
			Source: entry.Source{
				Uri:  "https://www.sglavoie.com/posts/2019/11/10/using-dmenu-to-optimize-common-tasks/",
				Text: "Sébastien Lavoie",
			},
			Added: entry.Date{Time: time.Now()},
			Read:  entry.Date{Time: time.Now()},
		}
		return m
	}
	m, err := entry.NewManager(serverfs.NativeFS(*storageDir))
	if err != nil {
		log.Fatal(err)
	}
	return m
}

func getServer() *server.Server {
	m := getEntryManager()
	if *allowLocal {
		log.Print("Serving from local directories")
		return server.NewFs(paths.Default, m, os.DirFS("static"), os.DirFS("dynamic"))
	} else {
		log.Print("Serving from embedded files")
		return server.New(paths.Default, m)
	}
}

func getListener() net.Listener {
	var ln net.Listener
	var err error
	if *useTsNet {
		s := &tsnet.Server{
			Hostname: "reading-list",
		}
		ln, err = s.Listen("tcp", *addr)
	} else {
		ln, err = net.Listen("tcp", *addr)
	}
	if err != nil {
		log.Fatal(err)
	}

	return tls.NewListener(ln, &tls.Config{
		GetCertificate: tailscale.GetCertificate,
	})
}

func main() {
	flag.Parse()
	logSettings, err := serverLog.Settings()
	if err != nil {
		log.Fatal(err)
	}

	srv := getServer()
	ln := getListener()

	log.Fatal(http.Serve(ln, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if *useTsNet {
			// Check Tailscale authentication
			who, err := tailscale.WhoIs(r.Context(), r.RemoteAddr)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			if logSettings.TailscaleIdentity {
				log.Printf("Tailscale-authenticated request from %s at %s", who.UserProfile.LoginName, who.Node.Name)
			}
		}
		if logSettings.RequestPath {
			log.Printf("Processing request for %v", r.URL)
		}
		srv.ServeHTTP(w, r)
	})))
}
