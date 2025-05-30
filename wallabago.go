package main

import "context"

type Id string

type Image struct {
	Name string
	Data []byte
}

type PageContent struct {
	Title  string
	Body   string // todo - separate paragraphs?
	Images []Image
}

type Url string

type Entry struct {
	Url         Url
	Review      string
	Annotations []Annotation
	Metadata    Metadata
	Favorite    bool
	Archived    bool
	Id          Id
	Content     PageContent
}

type Metadata struct {
	Author string
}

type AnnotationId string

type Annotation struct {
	Id   AnnotationId
	Text string
}

type PageRetrievealEngine interface {
	Retrive(context.Context, Url) (*PageContent, error)
}

type EntryStorage interface {
	Add(context.Context, Entry) (*Entry, error)
	Get(context.Context, Id) (*Entry, error)
	Update(context.Context, Entry) (*Entry, error)
	Delete(context.Context, Id) error
}

type Epub []byte // TODO

type EpubConversionEngine interface {
	ConvertToEpub(context.Context, Entry) (*Epub, error)
}

type ConversionEngine interface {
	EpubConversionEngine
}

type EntryManager struct {
	retrieval    PageRetrievealEngine
	entryStorage EntryStorage
}

type ReadabilityPageRetrievalEngine struct{}

func (e *ReadabilityPageRetrievalEngine) Retrive(context.Context, Url) (*PageContent, error) {
	panic("sike")
}

type SimpleEntryStorage struct{}

func (a *SimpleEntryStorage) Add(context.Context, Entry) (*Entry, error) {
	panic("sike")
}

func (a *SimpleEntryStorage) Get(context.Context, Id) (*Entry, error) {
	panic("sike")
}

func (a *SimpleEntryStorage) Update(context.Context, Entry) (*Entry, error) {
	panic("sike")
}

func (a *SimpleEntryStorage) Delete(context.Context, Id) error {
	panic("sike")
}

func NewEntryManager() *EntryManager {
	return &EntryManager{
		retrieval:    &ReadabilityPageRetrievalEngine{},
		entryStorage: &SimpleEntryStorage{},
	}
}

// Add retrieves the contents of the page and saves it
func (m *EntryManager) Add(ctx context.Context, entry Entry) (*Entry, error) {
	content, err := m.retrieval.Retrive(ctx, entry.Url)
	if err != nil {
		return nil, err
	}
	entry.Content = *content
	result, err := m.entryStorage.Add(ctx, entry)
	if err != nil {
		return nil, err
	}
	return result, nil
}
