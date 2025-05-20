package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"
)

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

const loggerKey = 777

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func LoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

func writeJSON(ctx context.Context, w http.ResponseWriter, value any) error {
	w.Header().Set("Content-Type", "application/json")
	logger := LoggerFromContext(ctx)
	body, err := json.Marshal(value)
	if err != nil {
		logger.Error("Failed to marshal JSON", "value", value, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return err
	}
	_, err = w.Write(body)
	if err != nil {
		logger.Error("Failed to write JSON body", "body", body, "error", err)
	} else {
		logger.Info("Sending JSON Response", "body", body)
	}
	return err
}

func setupServer() *http.Server {
	mux := http.DefaultServeMux
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) { // exact match
		slog.Info("Handling request", "page", r.URL, "method", r.Method)
		logger := LoggerFromContext(r.Context()).With(
			"method", r.Method,
			"url", r.URL,
			"host", r.Host,
			"remote_addr", r.RemoteAddr,
		)
		ctx := WithLogger(r.Context(), logger)
		start := time.Now()
		time.Sleep(1000 * time.Millisecond)
		duration := time.Since(start)
		w.Header().Add("Server-Timing", fmt.Sprintf(`handler;desc="time spend inside handler";dur=%f`, float64(duration.Microseconds())/1e3))
		writeJSON(ctx, w, &map[string]string{"hello": "world"})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { // anything else
		slog.Warn("Page not found", "page", r.URL)
		http.NotFound(w, r)
	})
	server := &http.Server{
		Handler: mux,
		Addr:    ":9999",
	}
	slog.Default().Info("Created server", "addr", server.Addr)
	return server
}

func main() {
	server := setupServer()
	log.Fatalln(server.ListenAndServe())
}
