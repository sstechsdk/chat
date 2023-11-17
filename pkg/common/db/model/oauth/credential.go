package oauth

type Credential struct {
	Base
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}
