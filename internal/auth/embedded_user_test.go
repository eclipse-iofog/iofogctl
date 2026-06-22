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

type mockEmbeddedAuthClient struct {
	loginEmail    string
	loginPassword string
	loginErr      error
	users         []client.AuthUserResponse
	created       []client.AuthUserCreateRequest
	loginCalls    int
	listCalls     int
}

func (m *mockEmbeddedAuthClient) bootstrapLogin(email, password string) error {
	m.loginCalls++
	m.loginEmail = email
	m.loginPassword = password
	if m.loginErr != nil {
		return m.loginErr
	}
	return nil
}

func (m *mockEmbeddedAuthClient) listAuthUsers() ([]client.AuthUserResponse, error) {
	m.listCalls++
	return m.users, nil
}

func (m *mockEmbeddedAuthClient) createAuthUser(req client.AuthUserCreateRequest) error {
	m.created = append(m.created, req)
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
	require.Zero(t, mock.loginCalls)
}

func TestEnsureIofogUserEmbeddedCreatesWhenMissing(t *testing.T) {
	t.Cleanup(resetEmbeddedAuthDeps)
	mock := &mockEmbeddedAuthClient{users: []client.AuthUserResponse{}}
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
	require.Equal(t, 1, mock.loginCalls)
	require.Equal(t, "admin", mock.loginEmail)
	require.Equal(t, "BootstrapTest12!", mock.loginPassword)
	require.Len(t, mock.created, 1)
	require.Equal(t, "user@domain.com", mock.created[0].Email)
	require.Equal(t, "LocalTest12!", mock.created[0].Password)
	require.Equal(t, []string{"admin"}, mock.created[0].Groups)
}

func TestEnsureIofogUserEmbeddedSkipsWhenExists(t *testing.T) {
	t.Cleanup(resetEmbeddedAuthDeps)
	mock := &mockEmbeddedAuthClient{
		users: []client.AuthUserResponse{{Email: "user@domain.com"}},
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
	require.Equal(t, 1, mock.loginCalls)
	require.Empty(t, mock.created)
}

func TestEnsureIofogUserEmbeddedUsesInteractivePassword(t *testing.T) {
	t.Cleanup(resetEmbeddedAuthDeps)
	mock := &mockEmbeddedAuthClient{users: []client.AuthUserResponse{}}
	newEmbeddedAuthClient = func(context.Context, string, string) (embeddedAuthClient, error) {
		return mock, nil
	}

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
	require.Equal(t, "LocalTest12!", mock.created[0].Password)
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
	readPasswordFn = readPasswordHidden
	isTerminalFn = term.IsTerminal
}
