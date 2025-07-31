package bootstrap

type ConditionName string

const (
	ConditionAdminCreated     ConditionName = "admin_created"
	ConditionWebClientCreated ConditionName = "web_client_created"
)

type Condition struct {
	Name      ConditionName
	Satisfied bool
}
