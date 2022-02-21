package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/cceckman/reading-list/server"
	"github.com/cceckman/reading-list/server/dynamic"
	"github.com/cceckman/reading-list/server/entry"
	serverfs "github.com/cceckman/reading-list/server/fs"
	serverLog "github.com/cceckman/reading-list/server/log"
	"github.com/cceckman/reading-list/server/paths"
	"github.com/cceckman/reading-list/server/static"

	"tailscale.com/client/tailscale"
	"tailscale.com/tsnet"
)

var (
	addr           = flag.String("addr", ":443", "Port or address:port to listen on")
	localTemplates = flag.String("localTemplates", "", "Serve templates from the local filesystem rather than embedded templates. Development only.")
	localStatic    = flag.String("localStatic", "", "Serve static content from the local filesystem rather than embedded templates. Development only.")
	storageDir     = flag.String("storage", "", "Directory to use for entry management. If empty, uses an in-memory entry store.")
	stateDir       = flag.String("state", "", "Directory to use for state management. If empty, uses the tsnet default.")
	tsNet          = flag.String("tsnet", "reading-list", "Advertise into a Tailscale network with the given name")
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
	render := dynamic.New()
	if *localTemplates != "" {
		log.Print("Using templates from filesystem")
		render = dynamic.NewFromFs(os.DirFS(*localTemplates))
	}
	var static fs.FS = static.Files
	if *localStatic != "" {
		log.Print("Using static content from filesystem")
		static = os.DirFS(*localStatic)
	}

	return server.New(paths.Default, m, render, static)
}

func getListener() net.Listener {
	var ln net.Listener
	var err error
	if *tsNet != "" {
		s := &tsnet.Server{
			Hostname: *tsNet,
			Dir:      *stateDir,
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

// Ensure the directory exists and has restrictive permissions.
func prepDirectory(name string) error {
	if info, err := os.Stat(name); errors.Is(err, os.ErrNotExist) {
		return os.MkdirAll(name, 0700)
	} else if err != nil {
		return fmt.Errorf("could not stat directory %q: %w", name, err)
	} else if !info.IsDir() {
		return fmt.Errorf("specified path %q is not a directory", name)
	}
	// We don't enforce restrictive permissions - just ensure creation sets them up.
	return nil
}

func main() {
	flag.Parse()
	logSettings, err := serverLog.Settings()
	if err != nil {
		log.Fatal(err)
	}
	if *stateDir != "" {
		// If no state directory is specified, allow tslib to create its own.
		if err := prepDirectory(*stateDir); err != nil {
			log.Fatal(err)
		}
	}
	if *storageDir != "" {
		// No storage directory --> use in-memory store.
		if err := prepDirectory(*storageDir); err != nil {
			log.Fatal(err)
		}
	}

	srv := getServer()
	ln := getListener()

	log.Fatal(http.Serve(ln, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if *tsNet != "" {
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
