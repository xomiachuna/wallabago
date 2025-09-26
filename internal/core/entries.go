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
	ID          int32
	URL         url.URL
	Title       string
	OwnerID     string
	Content     string
	SHA1        []byte
	CreatedAt   time.Time
	RetrievedAt *time.Time
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
