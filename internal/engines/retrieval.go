package engines

import (
	"context"
	"io"
	"net/http"
	"time"

	neturl "net/url"

	"github.com/andriihomiak/wallabago/internal/core"
	"github.com/cixtor/readability"
	"github.com/pkg/errors"
)

type SimpleReadabilityRetrievalEngine struct {
	userAgent string
	converter *readability.Readability
}

func (e *SimpleReadabilityRetrievalEngine) getPageContents(ctx context.Context, url neturl.URL) (io.ReadCloser, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), http.NoBody)
	request.Header.Set("User-Agent", e.userAgent)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, core.NewRetrievalError("failed to retrieve page", err)
	}
	if response.StatusCode >= 400 {
		response.Body.Close()
		return nil, core.NewRetrievalError("page returned error status", errors.Errorf("HTTP %d", response.StatusCode))
	}
	return response.Body, nil
}

func (e *SimpleReadabilityRetrievalEngine) parseEntry(page io.ReadCloser, url neturl.URL) (*core.BareEntry, error) {
	article, err := e.converter.Parse(page, url.String())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	now := time.Now()
	return &core.BareEntry{
		Title:       article.Title,
		URL:         url,
		RetrievedAt: now,
		Content:     article.TextContent,
	}, nil
}

func (e *SimpleReadabilityRetrievalEngine) RetrieveEntryByURL(ctx context.Context, url neturl.URL) (*core.BareEntry, error) {
	pageContents, err := e.getPageContents(ctx, url)
	if err != nil {
		return nil, err
	}
	entry, err := e.parseEntry(pageContents, url)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return entry, nil
}

func NewSimpleReadabilityRetrievalEngine(userAgent string) *SimpleReadabilityRetrievalEngine {
	return &SimpleReadabilityRetrievalEngine{
		userAgent: userAgent,
		converter: readability.New(),
	}
}
