package managers

import (
	"context"
	//nolint:gosec // the strength of sha1 is sufficient for our use-case
	"crypto/sha1"
	neturl "net/url"

	"github.com/andriihomiak/wallabago/internal/core"
	"github.com/andriihomiak/wallabago/internal/core/policy"
	"github.com/pkg/errors"
)

type EntryManager struct {
	authz     AuthorizationEngine
	entries   EntryStorage
	retrieval RetrievalEngine
}

func NewEntryManager(
	authz AuthorizationEngine,
	entries EntryStorage,
	retrieval RetrievalEngine,
) *EntryManager {
	return &EntryManager{
		authz:     authz,
		entries:   entries,
		retrieval: retrieval,
	}
}

func (em *EntryManager) AddEntry(ctx context.Context, accessToken core.AccessToken, newEntry core.NewEntry) (*core.Entry, error) {
	tx, err := em.authz.Begin(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	defer func() {
		rollbackOnError(ctx, err, tx.Rollback)
	}()

	err = em.authz.CheckPolicy(ctx, tx, accessToken.UserID, policy.Action{
		Subject:   policy.SubjectEntries,
		Operation: policy.OperationCreate,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	//nolint:gosec // sha1 is sufficient for our use-case
	urlHash := sha1.New().Sum([]byte(newEntry.URL.String()))

	exists, err := em.entries.EntryExistsBySHA1(ctx, tx, urlHash)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var entry *core.Entry

	if exists {
		entry, err = em.entries.GetEntryBySHA1(ctx, tx, urlHash)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	} else {
		entry, err = em.retrieval.RetrieveEntryByURL(ctx, newEntry.URL)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		err = em.entries.AddEntry(ctx, tx, *entry)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return entry, nil
}

func (em *EntryManager) EntryExists(ctx context.Context, accessToken core.AccessToken, url neturl.URL) (bool, error) {
	tx, err := em.authz.Begin(ctx)
	if err != nil {
		return false, errors.WithStack(err)
	}

	defer func() {
		rollbackOnError(ctx, err, tx.Rollback)
	}()

	// todo: check ownership?
	err = em.authz.CheckPolicy(ctx, tx, accessToken.UserID, policy.Action{
		Subject:   policy.SubjectEntries,
		Operation: policy.OperationRead,
	})
	if err != nil {
		return false, errors.WithStack(err)
	}

	//nolint:gosec // sha1 is sufficient for our use-case
	urlHash := sha1.New().Sum([]byte(url.String()))

	exists, err := em.entries.EntryExistsBySHA1(ctx, tx, urlHash)
	if err != nil {
		return false, errors.WithStack(err)
	}

	err = tx.Commit()
	if err != nil {
		return false, errors.WithStack(err)
	}

	return exists, nil
}
