package core

import "net/url"

type entry struct {
	URL   url.URL
	Title string
}

type Entry struct {
	entry
}

type NewEntry struct {
	entry
}
