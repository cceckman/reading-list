package paths

// HTTP path fragments for the URLs used by the server.
type Paths interface {
	// Path for the web share-target API.
	Share() string

	// Edit the entry with the given ID
	Edit(id string) string

	// Save an entry. The ID to save is passed via parameters (i.e. GET, POST, UPDATE...)
	Save() string

	// List available entries.
	List() string
}
