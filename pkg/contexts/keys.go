package contexts

type contextKey string

func (c contextKey) String() string {
	return "gosl context key " + string(c)
}

var (
	contextKeyAuthorizedUser = contextKey("auth-user")
	contextKeyRequestTime    = contextKey("req-time")
)
