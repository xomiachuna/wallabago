package core

import (
	"encoding/json"
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

// MarshalJSON implements custom JSON marshaling for Entry to serialize URL as a string.
func (e Entry) MarshalJSON() ([]byte, error) {
	type Alias Entry
	return json.Marshal(&struct {
		URL string `json:"url"`
		*Alias
	}{
		URL:   e.URL.String(),
		Alias: (*Alias)(&e),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for Entry to deserialize URL from a string.
func (e *Entry) UnmarshalJSON(data []byte) error {
	type Alias Entry
	aux := &struct {
		URL string `json:"url"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	parsedURL, err := url.Parse(aux.URL)
	if err != nil {
		return err
	}
	e.URL = *parsedURL
	return nil
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
