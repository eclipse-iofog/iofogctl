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
	createAuthUser(req client.AuthUserCreateRequest) (client.AuthUserResponse, error)
	resetAuthUserToken(userID string) (client.AuthUserResetTokenResponse, error)
	changePassword(req client.ChangePasswordRequest) error
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

	existing, found := findAuthUserByEmail(users, spec.User.Email)
	if found && !existing.MustChangePassword {
		return nil
	}

	var userID string
	if found {
		userID = existing.ID
	} else {
		tempPassword, err := generateTempPasswordFn()
		if err != nil {
			return fmt.Errorf("generate temporary password: %w", err)
		}

		created, err := clt.createAuthUser(client.AuthUserCreateRequest{
			Email:    spec.User.Email,
			Password: tempPassword,
			Groups:   []string{"admin"},
		})
		if err != nil {
			return err
		}
		userID = created.ID
		if userID == "" {
			return fmt.Errorf("controller did not return auth user id for %q", spec.User.Email)
		}
	}

	return finalizeEmbeddedUserPassword(clt, userID, password)
}

func finalizeEmbeddedUserPassword(clt embeddedAuthClient, userID, password string) error {
	tokenResp, err := clt.resetAuthUserToken(userID)
	if err != nil {
		return fmt.Errorf("reset auth user token: %w", err)
	}
	if tokenResp.ResetToken == "" {
		return fmt.Errorf("controller did not return reset token for auth user %q", userID)
	}

	if err := clt.changePassword(client.ChangePasswordRequest{
		ResetToken:  tokenResp.ResetToken,
		NewPassword: password,
	}); err != nil {
		return fmt.Errorf("set iofog user password: %w", err)
	}
	return nil
}

func findAuthUserByEmail(users []client.AuthUserResponse, email string) (*client.AuthUserResponse, bool) {
	for i := range users {
		if users[i].Email == email {
			return &users[i], true
		}
	}
	return nil, false
}
