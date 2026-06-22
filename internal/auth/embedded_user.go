package auth

import (
	"context"
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
)

const (
	AuthModeEmbedded = "embedded"
	AuthModeExternal = "external"
)

// EmbeddedAuthSpec holds inputs for EnsureIofogUserEmbedded.
type EmbeddedAuthSpec struct {
	Mode      string
	Bootstrap *rsc.AuthBootstrap
	User      *rsc.IofogUser
}

type embeddedAuthClient interface {
	bootstrapLogin(email, password string) error
	listAuthUsers() ([]client.AuthUserResponse, error)
	createAuthUser(req client.AuthUserCreateRequest) error
}

var newEmbeddedAuthClient = func(ctx context.Context, namespace, endpoint string) (embeddedAuthClient, error) {
	return newControllerAuthClient(ctx, namespace, endpoint, "")
}

// EnsureIofogUserEmbedded creates the iofogUser in embedded auth when missing (RFC §9).
func EnsureIofogUserEmbedded(ctx context.Context, namespace, endpoint string, spec EmbeddedAuthSpec) error {
	if spec.Mode != AuthModeEmbedded {
		return nil
	}
	if spec.Bootstrap == nil {
		return fmt.Errorf("embedded auth bootstrap credentials are required")
	}
	if spec.User == nil || spec.User.Email == "" {
		return fmt.Errorf("iofogUser email is required")
	}

	password := spec.User.GetRawPassword()
	if password == "" {
		var err error
		password, err = PromptIofogUserPassword(spec.User.Email)
		if err != nil {
			return err
		}
		spec.User.Password = password
		spec.User.EncodePassword()
	}

	clt, err := newEmbeddedAuthClient(ctx, namespace, endpoint)
	if err != nil {
		return err
	}

	if err := clt.bootstrapLogin(spec.Bootstrap.Username, spec.Bootstrap.Password); err != nil {
		return err
	}

	users, err := clt.listAuthUsers()
	if err != nil {
		return err
	}
	if authUserExists(users, spec.User.Email) {
		return nil
	}

	return clt.createAuthUser(client.AuthUserCreateRequest{
		Email:    spec.User.Email,
		Password: password,
		Groups:   []string{"admin"},
	})
}

func authUserExists(users []client.AuthUserResponse, email string) bool {
	for _, u := range users {
		if u.Email == email {
			return true
		}
	}
	return false
}
