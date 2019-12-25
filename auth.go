package cloudns

const (
	AuthTypeNone AuthType = iota
	AuthTypeUserID
	AuthTypeSubUserID
	AuthTypeSubUserName
)

type AuthRole int
type AuthType int

type Auth struct {
	Type        AuthType
	UserID      int
	SubUserID   int
	SubUserName string
	Password    string
}

func NewAuth() *Auth {
	return &Auth{Type: AuthTypeNone}
}

func (auth *Auth) GetParams() HttpParams {
	params := make(HttpParams)

	switch auth.Type {
	case AuthTypeNone:
		break
	case AuthTypeUserID:
		params["auth-id"] = auth.UserID
		params["auth-password"] = auth.Password
	case AuthTypeSubUserID:
		params["sub-auth-id"] = auth.SubUserID
		params["auth-password"] = auth.Password
	case AuthTypeSubUserName:
		params["sub-auth-user"] = auth.SubUserName
		params["auth-password"] = auth.Password
	default:
		panic("invalid authentication type")
	}

	return params
}
