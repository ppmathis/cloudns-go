package cloudns

// AuthType is an enumeration of the various ways of authenticating against the ClouDNS API
type AuthType int

// Enumeration values for AuthType
const (
	AuthTypeNone AuthType = iota
	AuthTypeUserID
	AuthTypeSubUserID
	AuthTypeSubUserName
)

// Auth provides methods for turning human-friendly credentials into API parameters
type Auth struct {
	Type        AuthType
	UserID      int
	SubUserID   int
	SubUserName string
	Password    string
}

// NewAuth instantiates an empty Auth which contains no credentials / AuthTypeNone
func NewAuth() *Auth {
	return &Auth{Type: AuthTypeNone}
}

// GetParams returns the correct API parameters for the ClouDNS API which should be provided by either query parameters
// (when using GET) or the POST body as JSON
func (auth *Auth) GetParams() HTTPParams {
	params := make(HTTPParams)

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

// getAllParamKeys returns all keys involved in authentication, which is being used to filter credentials out of
// automatically generated test fixtures
func (auth *Auth) getAllParamKeys() []string {
	return []string{"auth-id", "sub-auth-id", "sub-auth-user", "auth-password"}
}
