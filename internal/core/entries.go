package core

import "net/url"

type Entry struct {
	URL   url.URL
	Title string
}

type NewEntry struct {
	URL   url.URL
	Title string
}
