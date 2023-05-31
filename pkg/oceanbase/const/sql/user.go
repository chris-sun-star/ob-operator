package sql

const (
	CreateUser      = "create user if not exists ?"
	SetUserPassword = "alter user ? identified by ?"
	GrantPrivilege  = "grant ? on ? to ?"
)
