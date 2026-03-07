package core

import (
	"net/url"
	"time"
)

// BareEntry represents an entry without ownership metadata.
type BareEntry struct {
	URL         url.URL
	Title       string
	Content     string
	RetrievedAt time.Time
}

// Entry represents a downloaded article that is associated with an owner.
type Entry struct {
	ID    int32   `json:"id"`
	URL   url.URL `json:"url"`
	Title string  `json:"title"`
	// TODO: check for schema compliance with wallabag
	OwnerID     string     `json:"owner_id"`
	Content     string     `json:"content"`
	SHA1        []byte     `json:"-"`
	CreatedAt   time.Time  `json:"created_at"`
	RetrievedAt *time.Time `json:"retrieved_at"`
}

// NewEntry represents args for creating a new entry.
type NewEntry struct {
	URL   url.URL
	Title string
}

// EntryExistence represents the result of checking if an entry exists.
type EntryExistence struct {
	Exists bool `json:"exists"`
}
