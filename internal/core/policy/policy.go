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

type Action struct {
	Subject   Subject
	Operation Operation
}

var ErrForbidden = fmt.Errorf("forbidden")
