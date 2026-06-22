package auth

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/stretchr/testify/require"
	"golang.org/x/term"
)

const testTempPassword = "TempPass123!"

type mockEmbeddedAuthClient struct {
	logins          [][2]string
	loginErr        error
	users           []client.AuthUserResponse
	created         []client.AuthUserCreateRequest
	resetTokenCalls []string
	resetTokens     []client.AuthUserResetTokenResponse
	changedPassword []client.ChangePasswordRequest
	listCalls       int
}

func (m *mockEmbeddedAuthClient) bootstrapLogin(email, password string) error {
	m.logins = append(m.logins, [2]string{email, password})
	if m.loginErr != nil {
		return m.loginErr
	}
	return nil
}

func (m *mockEmbeddedAuthClient) listAuthUsers() ([]client.AuthUserResponse, error) {
	m.listCalls++
	return m.users, nil
}

func (m *mockEmbeddedAuthClient) createAuthUser(req client.AuthUserCreateRequest) (client.AuthUserResponse, error) {
	m.created = append(m.created, req)
	return client.AuthUserResponse{
		ID:                 "user-1",
		Email:              req.Email,
		Groups:             req.Groups,
		MustChangePassword: true,
	}, nil
}

func (m *mockEmbeddedAuthClient) resetAuthUserToken(userID string) (client.AuthUserResetTokenResponse, error) {
	m.resetTokenCalls = append(m.resetTokenCalls, userID)
	if len(m.resetTokens) > 0 {
		return m.resetTokens[0], nil
	}
	return client.AuthUserResetTokenResponse{ResetToken: "reset-token-1", ExpiresIn: 900}, nil
}

func (m *mockEmbeddedAuthClient) changePassword(req client.ChangePasswordRequest) error {
	m.changedPassword = append(m.changedPassword, req)
	return nil
}

func TestEnsureIofogUserEmbeddedSkipsExternal(t *testing.T) {
	t.Cleanup(resetEmbeddedAuthDeps)
	mock := &mockEmbeddedAuthClient{}
	newEmbeddedAuthClient = func(context.Context, string, string) (embeddedAuthClient, error) {
		return mock, nil
	}

	err := EnsureIofogUserEmbedded(context.Background(), "default", "https://controller.example.com", EmbeddedAuthSpec{
		Mode: AuthModeExternal,
		User: &rsc.IofogUser{Email: "user@domain.com"},
	})
	require.NoError(t, err)
	require.Empty(t, mock.logins)
}

func TestEnsureIofogUserEmbeddedCreatesWhenMissing(t *testing.T) {
	t.Cleanup(resetEmbeddedAuthDeps)
	mock := &mockEmbeddedAuthClient{users: []client.AuthUserResponse{}}
	newEmbeddedAuthClient = func(context.Context, string, string) (embeddedAuthClient, error) {
		return mock, nil
	}
	generateTempPasswordFn = func() (string, error) { return testTempPassword, nil }

	user := &rsc.IofogUser{Email: "user@domain.com", Password: "LocalTest12!"}
	user.EncodePassword()

	err := EnsureIofogUserEmbedded(context.Background(), "default", "https://controller.example.com", EmbeddedAuthSpec{
		Mode:      AuthModeEmbedded,
		Bootstrap: &rsc.AuthBootstrap{Username: "admin", Password: "BootstrapTest12!"},
		User:      user,
	})
	require.NoError(t, err)
	require.Len(t, mock.logins, 1)
	require.Equal(t, "admin", mock.logins[0][0])
	require.Equal(t, "BootstrapTest12!", mock.logins[0][1])
	require.Len(t, mock.created, 1)
	require.Equal(t, "user@domain.com", mock.created[0].Email)
	require.Equal(t, testTempPassword, mock.created[0].Password)
	require.Equal(t, []string{"admin"}, mock.created[0].Groups)
	require.Equal(t, []string{"user-1"}, mock.resetTokenCalls)
	require.Len(t, mock.changedPassword, 1)
	require.Equal(t, "reset-token-1", mock.changedPassword[0].ResetToken)
	require.Equal(t, "LocalTest12!", mock.changedPassword[0].NewPassword)
	require.Empty(t, mock.changedPassword[0].CurrentPassword)
}

func TestEnsureIofogUserEmbeddedSkipsWhenReady(t *testing.T) {
	t.Cleanup(resetEmbeddedAuthDeps)
	mock := &mockEmbeddedAuthClient{
		users: []client.AuthUserResponse{{
			ID:                 "user-1",
			Email:              "user@domain.com",
			MustChangePassword: false,
		}},
	}
	newEmbeddedAuthClient = func(context.Context, string, string) (embeddedAuthClient, error) {
		return mock, nil
	}

	user := &rsc.IofogUser{Email: "user@domain.com", Password: "LocalTest12!"}
	user.EncodePassword()

	err := EnsureIofogUserEmbedded(context.Background(), "default", "https://controller.example.com", EmbeddedAuthSpec{
		Mode:      AuthModeEmbedded,
		Bootstrap: &rsc.AuthBootstrap{Username: "admin", Password: "BootstrapTest12!"},
		User:      user,
	})
	require.NoError(t, err)
	require.Len(t, mock.logins, 1)
	require.Empty(t, mock.created)
	require.Empty(t, mock.resetTokenCalls)
	require.Empty(t, mock.changedPassword)
}

func TestEnsureIofogUserEmbeddedFinalizesExistingTempUser(t *testing.T) {
	t.Cleanup(resetEmbeddedAuthDeps)
	mock := &mockEmbeddedAuthClient{
		users: []client.AuthUserResponse{{
			ID:                 "existing-user",
			Email:              "user@domain.com",
			MustChangePassword: true,
		}},
	}
	newEmbeddedAuthClient = func(context.Context, string, string) (embeddedAuthClient, error) {
		return mock, nil
	}

	user := &rsc.IofogUser{Email: "user@domain.com", Password: "LocalTest12!"}
	user.EncodePassword()

	err := EnsureIofogUserEmbedded(context.Background(), "default", "https://controller.example.com", EmbeddedAuthSpec{
		Mode:      AuthModeEmbedded,
		Bootstrap: &rsc.AuthBootstrap{Username: "admin", Password: "BootstrapTest12!"},
		User:      user,
	})
	require.NoError(t, err)
	require.Empty(t, mock.created)
	require.Equal(t, []string{"existing-user"}, mock.resetTokenCalls)
	require.Len(t, mock.changedPassword, 1)
	require.Equal(t, "LocalTest12!", mock.changedPassword[0].NewPassword)
}

func TestEnsureIofogUserEmbeddedUsesInteractivePassword(t *testing.T) {
	t.Cleanup(resetEmbeddedAuthDeps)
	mock := &mockEmbeddedAuthClient{users: []client.AuthUserResponse{}}
	newEmbeddedAuthClient = func(context.Context, string, string) (embeddedAuthClient, error) {
		return mock, nil
	}
	generateTempPasswordFn = func() (string, error) { return testTempPassword, nil }

	promptCalls := 0
	readPasswordFn = func(prompt string) (string, error) {
		promptCalls++
		if strings.Contains(prompt, "Confirm") {
			return "LocalTest12!", nil
		}
		return "LocalTest12!", nil
	}
	isTerminalFn = func(int) bool { return true }

	user := &rsc.IofogUser{Email: "user@domain.com"}
	err := EnsureIofogUserEmbedded(context.Background(), "default", "https://controller.example.com", EmbeddedAuthSpec{
		Mode:      AuthModeEmbedded,
		Bootstrap: &rsc.AuthBootstrap{Username: "admin", Password: "BootstrapTest12!"},
		User:      user,
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, promptCalls, 2)
	require.NotEmpty(t, user.GetRawPassword())
	require.Len(t, mock.created, 1)
	require.Equal(t, testTempPassword, mock.created[0].Password)
	require.Len(t, mock.changedPassword, 1)
	require.Equal(t, "LocalTest12!", mock.changedPassword[0].NewPassword)
}

func TestEnsureIofogUserEmbeddedPropagatesLoginError(t *testing.T) {
	t.Cleanup(resetEmbeddedAuthDeps)
	newEmbeddedAuthClient = func(context.Context, string, string) (embeddedAuthClient, error) {
		return &mockEmbeddedAuthClient{loginErr: errors.New("login failed")}, nil
	}

	user := &rsc.IofogUser{Email: "user@domain.com", Password: "LocalTest12!"}
	user.EncodePassword()

	err := EnsureIofogUserEmbedded(context.Background(), "default", "https://controller.example.com", EmbeddedAuthSpec{
		Mode:      AuthModeEmbedded,
		Bootstrap: &rsc.AuthBootstrap{Username: "admin", Password: "BootstrapTest12!"},
		User:      user,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "login failed")
}

func resetEmbeddedAuthDeps() {
	newEmbeddedAuthClient = func(ctx context.Context, namespace, endpoint string) (embeddedAuthClient, error) {
		return newControllerAuthClient(ctx, namespace, endpoint, "")
	}
	generateTempPasswordFn = generateTempPassword
	readPasswordFn = readPasswordHidden
	isTerminalFn = term.IsTerminal
}
