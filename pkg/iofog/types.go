package iofog

type User struct {
	Name     string `json:"firstName"`
	Surname  string `json:"lastName"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ControllerStatus struct {
	Status      string `json:"status"`
	UptimeMsUTC int64  `json:"timestamp"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"accessToken"`
}

type CreateAgentRequest struct {
	Name    string `json:"name"`
	FogType int32  `json:"fogType"`
}

type CreateAgentResponse struct {
	UUID string
}

type GetAgentProvisionKeyResponse struct {
	Key         string `json:"key"`
	ExpireMsUTC int64  `json:"expirationTime"`
}
