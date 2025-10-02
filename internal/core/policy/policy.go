package policy

import "fmt"

type Subject string

var (
	SubjectEntries Subject = "entries"
	SubjectUsers   Subject = "users"
)

type Operation string

var (
	OperationCreate Operation = "create"
	OperationRead   Operation = "read"
	OperationUpdate Operation = "update"
	OperationDelete Operation = "delete"
	OperationManage Operation = "manage"
	OperationExport Operation = "export"
)

type Scope string

var (
	ScopeNone   Scope = "none"
	ScopeOwn    Scope = "own"
	ScopeGlobal Scope = "global"
)

type Role string

var (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type Resource struct {
	ID      any    // Resource ID (type depends on resource: int32 for entries, string for users, etc.)
	OwnerID string // Optional - if empty, authorization engine will look it up
}

type Action struct {
	Subject   Subject
	Operation Operation
	Resource  *Resource // nil for permission checks without specific resource context
}

var ErrForbidden = fmt.Errorf("forbidden")
