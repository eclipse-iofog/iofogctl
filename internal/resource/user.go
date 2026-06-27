package resource

import (
	"encoding/base64"
)

// IofogUser contains information about users registered against a controller
type IofogUser struct {
	Name         string `yaml:"name,omitempty"`
	Surname      string `yaml:"surname,omitempty"`
	Email        string `yaml:"email,omitempty"`
	Password     string `yaml:"password,omitempty"`
	AccessToken  string `yaml:"accessToken,omitempty"`
	RefreshToken string `yaml:"refreshToken,omitempty"`
}

func (user *IofogUser) EncodePassword() {
	user.Password = encodeBase64(user.Password)
}

func (user IofogUser) GetRawPassword() string {
	buf, err := base64.StdEncoding.DecodeString(user.Password)
	if err != nil {
		return user.Password
	}
	return string(buf)
}

func encodeBase64(raw string) string {
	return base64.StdEncoding.EncodeToString([]byte(raw))
}
