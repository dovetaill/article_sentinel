package identity

type PrincipalKind string

const (
	PrincipalAdmin  PrincipalKind = "admin"
	PrincipalMember PrincipalKind = "member"
)

type Principal struct {
	Kind     PrincipalKind `json:"kind"`
	UserID   uint          `json:"user_id"`
	Username string        `json:"username"`
	Role     string        `json:"role"`
	Status   string        `json:"status"`
}

func PrincipalFromActor(actor Actor) Principal {
	return Principal{
		Kind:     PrincipalKind(actor.Role),
		UserID:   actor.ID,
		Username: actor.Username,
		Role:     actor.Role,
		Status:   actor.Status,
	}
}
