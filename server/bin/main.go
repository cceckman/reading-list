package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/cceckman/reading-list/server"
	"github.com/cceckman/reading-list/server/entry"
	serverLog "github.com/cceckman/reading-list/server/log"
	"github.com/cceckman/reading-list/server/paths"

	"tailscale.com/client/tailscale"
	"tailscale.com/tsnet"
)

var (
	addr       = flag.String("addr", ":443", "Port or address:port to listen on")
	allowLocal = flag.Bool("allowLocal", false, "Allow serving from the local static/ directory rather than embedded content. Development only.")
	useTsNet   = flag.Bool("tsnet", true, "Connect directly to Tailscale via tsnet")
)

func main() {
	flag.Parse()

	var m entry.TestEntryManager

	var s *server.Server
	if *allowLocal {
		log.Print("Serving from local directories")
		s = server.NewFs(paths.Default, &m, os.DirFS("static"), os.DirFS("dynamic"))
	} else {
		log.Print("Serving from embedded files")
		s = server.New(paths.Default, &m)
	}

	logSettings, err := serverLog.Settings()
	if err != nil {
		log.Fatal(err)
	}

	var ln net.Listener
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

	tls := tls.NewListener(ln, &tls.Config{
		GetCertificate: tailscale.GetCertificate,
	})

	log.Fatal(http.Serve(tls, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		s.ServeHTTP(w, r)
	})))
}
