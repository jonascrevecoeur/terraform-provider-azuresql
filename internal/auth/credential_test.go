package auth

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func clearEnv(t *testing.T) {
	for _, list := range [][]string{EnvOIDCToken, EnvOIDCTokenFilePath, EnvOIDCRequestURL, EnvOIDCRequestToken, EnvServiceConnectionID} {
		for _, name := range list {
			t.Setenv(name, "")
		}
	}
}

// The Azure DevOps TerraformTask sets these variables for a service connection
// using workload identity federation; SYSTEM_OIDCREQUESTURI is set for every job.
func TestFromEnvironmentAzureDevOps(t *testing.T) {
	clearEnv(t)
	t.Setenv("SYSTEM_OIDCREQUESTURI", "https://dev.azure.com/org/")
	t.Setenv("ARM_OIDC_REQUEST_TOKEN", "system-token")
	t.Setenv("ARM_OIDC_AZURE_SERVICE_CONNECTION_ID", "sc")

	c := Config{}.FromEnvironment()

	if c.OIDCRequestURL != "https://dev.azure.com/org/" || c.OIDCRequestToken != "system-token" || c.ServiceConnectionID != "sc" {
		t.Errorf("unexpected config %+v", c)
	}
}

func TestFromEnvironmentGitHubActions(t *testing.T) {
	clearEnv(t)
	t.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", "https://token.actions.example")
	t.Setenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN", "request-token")

	c := Config{}.FromEnvironment()

	if c.OIDCRequestURL != "https://token.actions.example" || c.OIDCRequestToken != "request-token" || c.ServiceConnectionID != "" {
		t.Errorf("unexpected config %+v", c)
	}
}

func TestFromEnvironmentPrecedence(t *testing.T) {
	clearEnv(t)
	t.Setenv("ARM_OIDC_REQUEST_URL", "arm")
	t.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", "gh")
	t.Setenv("SYSTEM_OIDCREQUESTURI", "ado")
	t.Setenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN", "gh-token")
	t.Setenv("SYSTEM_ACCESSTOKEN", "system")
	t.Setenv("ARM_ADO_PIPELINE_SERVICE_CONNECTION_ID", "new")
	t.Setenv("ARM_OIDC_AZURE_SERVICE_CONNECTION_ID", "old")

	c := Config{}.FromEnvironment()

	if c.OIDCRequestURL != "arm" || c.OIDCRequestToken != "gh-token" || c.ServiceConnectionID != "new" {
		t.Errorf("unexpected config %+v", c)
	}
}

func TestNewCredentialDisabled(t *testing.T) {
	// OIDC token sources alone do not configure an authentication method
	cred, err := NewCredential(context.Background(), Config{OIDCRequestURL: "https://x", OIDCRequestToken: "t"})
	if err != nil || cred != nil {
		t.Errorf("expected nil credential and nil error, got %v, %v", cred, err)
	}
}

func TestNewCredentialEnabled(t *testing.T) {
	tests := map[string]Config{
		"client secret":          {ClientID: "client", TenantID: "tenant", ClientSecret: "secret"},
		"managed identity":       {UseMSI: true},
		"user-assigned identity": {UseMSI: true, ClientID: "client"},
		"azure devops":           {UseOIDC: true, ClientID: "client", TenantID: "tenant", OIDCRequestURL: "https://dev.azure.com/org/", OIDCRequestToken: "t", ServiceConnectionID: "sc"},
		"token":                  {UseOIDC: true, ClientID: "client", TenantID: "tenant", OIDCToken: "jwt"},
		"github actions":         {UseOIDC: true, ClientID: "client", TenantID: "tenant", OIDCRequestURL: "https://token.actions.example", OIDCRequestToken: "t"},
	}

	for name, config := range tests {
		t.Run(name, func(t *testing.T) {
			cred, err := NewCredential(context.Background(), config)
			if err != nil || cred == nil {
				t.Errorf("expected a credential, got %v, %v", cred, err)
			}
		})
	}
}

func TestNewCredentialValidation(t *testing.T) {
	tests := map[string]Config{
		"client id without method": {ClientID: "client", TenantID: "tenant"},
		"secret without tenant":    {ClientID: "client", ClientSecret: "secret"},
		"secret without client id": {TenantID: "tenant", ClientSecret: "secret"},
		"invalid certificate":      {ClientID: "client", TenantID: "tenant", ClientCertificate: "not base64"},
		"missing certificate file": {ClientID: "client", TenantID: "tenant", ClientCertificatePath: "/does/not/exist.pfx"},
		"missing client id":        {UseOIDC: true, TenantID: "tenant", OIDCRequestURL: "https://x", OIDCRequestToken: "t"},
		"missing tenant id":        {UseOIDC: true, ClientID: "client", OIDCRequestURL: "https://x", OIDCRequestToken: "t"},
		"missing request url":      {UseOIDC: true, ClientID: "client", TenantID: "tenant", OIDCRequestToken: "t"},
		"unreadable token file":    {UseOIDC: true, ClientID: "client", TenantID: "tenant", OIDCTokenFilePath: "/does/not/exist"},
		"missing request token":    {UseOIDC: true, ClientID: "client", TenantID: "tenant", OIDCRequestURL: "https://x"},
	}

	for name, config := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := NewCredential(context.Background(), config); err == nil {
				t.Error("expected an error")
			}
		})
	}
}

func TestAuthorizerPerScope(t *testing.T) {
	cred, err := NewCredential(context.Background(), Config{UseOIDC: true, ClientID: "client", TenantID: "tenant", OIDCRequestURL: "https://x", OIDCRequestToken: "t"})
	if err != nil {
		t.Fatal(err)
	}
	c := cred.(*Credential)

	// the scope announced by the SQL server and the default scope share one authorizer
	a1, err := c.authorizer(context.Background(), "https://database.windows.net//.default")
	if err != nil {
		t.Fatal(err)
	}
	a2, _ := c.authorizer(context.Background(), "https://database.windows.net/.default")
	a3, _ := c.authorizer(context.Background(), "https://management.azure.com/.default")

	if a1 != a2 {
		t.Error("expected the same authorizer for the same resource")
	}
	if a1 == a3 {
		t.Error("expected a separate authorizer per resource")
	}
}

// A token issued by any OIDC provider (e.g. a Kubernetes service account token
// or a Keycloak token) can be supplied directly or as a file.
func TestFromEnvironmentToken(t *testing.T) {
	clearEnv(t)
	t.Setenv("ARM_OIDC_TOKEN", "jwt")
	t.Setenv("ARM_OIDC_TOKEN_FILE_PATH", "/var/run/secrets/token")

	c := Config{}.FromEnvironment()

	if c.OIDCToken != "jwt" || c.OIDCTokenFilePath != "/var/run/secrets/token" {
		t.Errorf("unexpected config %+v", c)
	}
}

func TestOIDCTokenFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "token")
	if err := os.WriteFile(path, []byte("file-jwt\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if got, err := oidcToken(Config{OIDCTokenFilePath: path}); err != nil || got != "file-jwt" {
		t.Errorf("got %q, %v; want file-jwt", got, err)
	}
	if got, err := oidcToken(Config{OIDCToken: "file-jwt", OIDCTokenFilePath: path}); err != nil || got != "file-jwt" {
		t.Errorf("matching token and file: got %q, %v", got, err)
	}
	if _, err := oidcToken(Config{OIDCToken: "other", OIDCTokenFilePath: path}); err == nil {
		t.Error("expected an error for a token that differs from the file contents")
	}

	cred, err := NewCredential(context.Background(), Config{UseOIDC: true, ClientID: "client", TenantID: "tenant", OIDCTokenFilePath: path})
	if err != nil || cred == nil {
		t.Errorf("expected a credential, got %v, %v", cred, err)
	}
}
