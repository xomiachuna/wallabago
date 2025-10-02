package engines

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/andriihomiak/wallabago/internal/core/policy"
)

type HierarchicalAuthorizationEngine struct {
	rbacStorage RBACStorage
}

func NewRBACAuthorizationEngine(
	rbacStorage RBACStorage,
) *HierarchicalAuthorizationEngine {
	return &HierarchicalAuthorizationEngine{
		rbacStorage: rbacStorage,
	}
}

func (e *HierarchicalAuthorizationEngine) Begin(ctx context.Context) (*sql.Tx, error) {
	return e.rbacStorage.Begin(ctx)
}

func (e *HierarchicalAuthorizationEngine) CheckPolicy(ctx context.Context, tx *sql.Tx, userID string, action policy.Action) error {
	// 1. Get the maximum scope the user has for this action
	scope, err := e.rbacStorage.GetUserMaxScope(ctx, tx, userID, action.Subject, action.Operation)
	if err != nil {
		return err
	}

	// 2. Check global scope (highest level - allows everything)
	if scope == policy.ScopeGlobal {
		slog.DebugContext(ctx, "authorization granted: global scope",
			"userID", userID,
			"subject", action.Subject,
			"operation", action.Operation)
		return nil
	}

	// 3. Check own scope (requires owner match)
	if scope == policy.ScopeOwn {
		// If no resource specified, user has the permission in general
		if action.Resource == nil {
			slog.DebugContext(ctx, "authorization granted: own scope, no specific resource",
				"userID", userID,
				"subject", action.Subject,
				"operation", action.Operation)
			return nil
		}

		// Resource specified - need to verify ownership
		resourceOwnerID := action.Resource.OwnerID

		// If owner not provided, look it up based on resource type
		if resourceOwnerID == "" {
			var lookupErr error
			resourceOwnerID, lookupErr = e.lookupResourceOwner(ctx, tx, action.Subject, action.Resource.ID)
			if lookupErr != nil {
				slog.WarnContext(ctx, "failed to lookup resource owner",
					"subject", action.Subject,
					"resourceID", action.Resource.ID,
					"error", lookupErr)
				return lookupErr
			}
		}

		// Check if user owns the resource
		if resourceOwnerID == userID {
			slog.DebugContext(ctx, "authorization granted: own scope, owner match",
				"userID", userID,
				"subject", action.Subject,
				"operation", action.Operation,
				"resourceID", action.Resource.ID)
			return nil
		}

		// User doesn't own the resource
		slog.WarnContext(ctx, "authorization denied: own scope, owner mismatch",
			"userID", userID,
			"subject", action.Subject,
			"operation", action.Operation,
			"resourceID", action.Resource.ID,
			"resourceOwner", resourceOwnerID)
		return policy.ErrForbidden
	}

	// 4. No sufficient permission (none or unknown scope)
	slog.WarnContext(ctx, "authorization denied: insufficient scope",
		"userID", userID,
		"subject", action.Subject,
		"operation", action.Operation,
		"scope", scope)
	return policy.ErrForbidden
}

func (e *HierarchicalAuthorizationEngine) lookupResourceOwner(ctx context.Context, tx *sql.Tx, subject policy.Subject, resourceID any) (string, error) {
	switch subject {
	case policy.SubjectEntries:
		entryID, ok := resourceID.(int32)
		if !ok {
			slog.ErrorContext(ctx, "invalid entry ID type",
				"resourceID", resourceID,
				"expectedType", "int32")
			return "", policy.ErrForbidden
		}
		return e.rbacStorage.GetEntryOwnerID(ctx, tx, entryID)

	case policy.SubjectUsers:
		// Users own themselves
		userID, ok := resourceID.(string)
		if !ok {
			slog.ErrorContext(ctx, "invalid user ID type",
				"resourceID", resourceID,
				"expectedType", "string")
			return "", policy.ErrForbidden
		}
		return userID, nil

	default:
		slog.ErrorContext(ctx, "unknown subject type for owner lookup",
			"subject", string(subject))
		return "", policy.ErrForbidden
	}
}

type RBACStorage interface {
	Begin(ctx context.Context) (*sql.Tx, error)

	// Permission queries
	GetUserMaxScope(ctx context.Context, tx *sql.Tx, userID string, subject policy.Subject, operation policy.Operation) (policy.Scope, error)

	// Resource owner lookups
	GetEntryOwnerID(ctx context.Context, tx *sql.Tx, entryID int32) (string, error)
}
