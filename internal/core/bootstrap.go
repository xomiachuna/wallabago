package core

type ConditionName string

const (
	ConditionAdminCreated     ConditionName = "admin_created"
	ConditionWebClientCreated ConditionName = "web_client_created"
)

type Condition struct {
	Name      ConditionName
	Satisfied bool
}

type BootstrapAdminCredentials struct {
	Username string
	Password string
	Email    string
}
