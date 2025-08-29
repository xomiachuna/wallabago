// Place to keep track of common dependencies used in multiple managers
package managers

import (
	"context"
	"database/sql"
	"net/url"

	"github.com/andriihomiak/wallabago/internal/core"
	"github.com/andriihomiak/wallabago/internal/core/policy"
)

type AuthorizationEngine interface {
	CheckPolicy(ctx context.Context, tx *sql.Tx, userID string, action policy.Action) error

	transactionStarter
}

type EntryStorage interface {
	EntryExistsBySHA1(ctx context.Context, tx *sql.Tx, hash []byte) (bool, error)
	GetEntryBySHA1(ctx context.Context, tx *sql.Tx, hash []byte) (*core.Entry, error)
	AddEntry(ctx context.Context, tx *sql.Tx, entry core.Entry) error

	transactionStarter
}

type RetrievalEngine interface {
	RetrieveEntryByURL(ctx context.Context, url url.URL) (*core.Entry, error)
}
