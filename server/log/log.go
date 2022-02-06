// Logging verbosity controls.
package log

import (
	"flag"
	"fmt"
	"strings"
)

var (
	logmodule = flag.String("logmodule", "", "Events to log. Comma-separated string, or 'all'.")
)

type LogSettings struct {
	TailscaleIdentity bool
	RequestPath       bool
}

// Parses the log settings from flags.
// Returns an error if there are any unrecognized settings;
// the (parsed) settings are still returned even if there is an error.
// This means it's up to the caller to error or return!
func Settings() (LogSettings, error) {
	flag.Parse()
	modules := strings.Split(*logmodule, ",")
	var unrecognizedModules []string

	settings := LogSettings{}
	for _, mod := range modules {
		switch mod {
		case "all":
			settings.RequestPath = true
			settings.TailscaleIdentity = true
		case "requestPath":
			settings.RequestPath = true
		case "tsIdent":
			settings.TailscaleIdentity = true
		default:
			unrecognizedModules = append(unrecognizedModules, mod)
		}
	}
	var e error
	if len(unrecognizedModules) != 0 {
		e = fmt.Errorf("unrecognized log modules: %+v", unrecognizedModules)
	}
	return settings, e
}