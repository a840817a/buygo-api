package interceptor

//go:generate go run ../../../scripts/gen_auth_policy.go

type accessLevel int

const (
	accessAuthenticated accessLevel = iota
	accessPublic
	accessSysAdmin
)

func procedureAccess(procedure string) accessLevel {
	level, ok := procedureAccessPolicy[procedure]
	if !ok {
		return accessAuthenticated
	}
	return level
}
