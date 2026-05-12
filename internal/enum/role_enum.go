package enum

type Role int

const (
	RoleUser Role = iota
	RoleModerator
	RoleAdmin
)

var roleToString = map[Role]string{
	RoleUser:      "user",
	RoleModerator: "moderator",
	RoleAdmin:     "admin",
}

func (r Role) String() string {
	return roleToString[r]
}

func RoleFromString(s string) (Role, bool) {
	for k, v := range roleToString {
		if v == s {
			return k, true
		}
	}
	return RoleUser, false // default to RoleUser if not found
}
