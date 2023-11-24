package oauth

type Credential struct {
	Base
	AuthResp
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type AuthResp struct {
	AppId            string `json:"appId"`
	AppSecret        string `json:"appSecret"`
	CallbackUrl      string `json:"callbackUrl"`
	ServerPublicKey  string `json:"serverPublicKey"`
	ServerPrivateKey string `json:"serverPrivateKey"`
	ClientPublicKey  string `json:"clientPublicKey"`
	AppName          string `json:"appName"`
}

//type AuthModel struct {
//	AppId            string `json:"appId"`
//	AppSecret        string `json:"appSecret"`
//	CallbackUrl      string `json:"callbackUrl"`
//	ServerPublicKey  string `json:"serverPublicKey"`
//	ServerPrivateKey string `json:"serverPrivateKey"`
//	ClientPublicKey  string `json:"clientPublicKey"`
//	AppName          string `json:"appName"`
//}

type AuthClient struct {
	ID               string
	Secret           string
	Domain           string
	UserID           string
	CallbackUrl      string
	ServerPublicKey  string
	ServerPrivateKey string
	ClientPublicKey  string
	AppName          string
}

// GetID client id
func (c *AuthClient) GetID() string {
	return c.ID
}

// GetSecret client domain
func (c *AuthClient) GetSecret() string {
	return c.Secret
}

// GetDomain client domain
func (c *AuthClient) GetDomain() string {
	return c.Domain
}

// GetUserID user id
func (c *AuthClient) GetUserID() string {
	return c.UserID
}

// GetCallbackUrl client id
func (c *AuthClient) GetCallbackUrl() string {
	return c.CallbackUrl
}

// GetServerPublicKey client domain
func (c *AuthClient) GetServerPublicKey() string {
	return c.ServerPublicKey
}

// GetServerPrivateKey client domain
func (c *AuthClient) GetServerPrivateKey() string {
	return c.ServerPrivateKey
}

// GetClientPublicKey user id
func (c *AuthClient) GetClientPublicKey() string {
	return c.ClientPublicKey
}

// GetAppName user id
func (c *AuthClient) GetAppName() string {
	return c.AppName
}
