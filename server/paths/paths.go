package paths

// HTTP path fragments for the URLs used by the server.
type Paths interface {
	// Path for the web share-target API.
	Share() string

	// Edit an entry
	Edit() string

	// Save an entry. The ID to save is passed via parameters (i.e. GET, POST, UPDATE...)
	Save() string

	// List available entries.
	List() string
}

var Default Paths = &defaults{}

type defaults struct{}

func (*defaults) Share() string {
	return "/share"
}

func (*defaults) Edit() string {
	return "/edit"
}

func (*defaults) Save() string {
	return "/save"
}

func (*defaults) List() string {
	return "/list"
}
