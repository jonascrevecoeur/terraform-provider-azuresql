// Package auth builds the Azure credential used by the provider when an
// authentication method is configured: client secret, client certificate,
// OIDC (workload identity federation) or managed identity.
//
// Authentication is delegated to github.com/hashicorp/go-azure-sdk/sdk/auth.
// For OIDC, the token is either
//
//   - requested from a token endpoint (Azure DevOps pipelines, GitHub Actions)
//     whenever a new Entra ID token is needed, so it is refreshed during long
//     running plans and applies, or
//   - supplied directly or as a file by any issuer (e.g. Kubernetes or
//     Keycloak), and read once when the provider is configured.
//
// Without a configured method, the provider keeps using the Azure default
// credential chain of the SQL driver.
package auth

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	sdkauth "github.com/hashicorp/go-azure-sdk/sdk/auth"
	"github.com/hashicorp/go-azure-sdk/sdk/environments"
)

// Environment variables providing the OIDC token, in order of precedence.
var (
	EnvOIDCToken           = []string{"ARM_OIDC_TOKEN"}
	EnvOIDCTokenFilePath   = []string{"ARM_OIDC_TOKEN_FILE_PATH"}
	EnvOIDCRequestURL      = []string{"ARM_OIDC_REQUEST_URL", "ACTIONS_ID_TOKEN_REQUEST_URL", "SYSTEM_OIDCREQUESTURI"}
	EnvOIDCRequestToken    = []string{"ARM_OIDC_REQUEST_TOKEN", "ACTIONS_ID_TOKEN_REQUEST_TOKEN", "SYSTEM_ACCESSTOKEN"}
	EnvServiceConnectionID = []string{"ARM_ADO_PIPELINE_SERVICE_CONNECTION_ID", "ARM_OIDC_AZURE_SERVICE_CONNECTION_ID"}
)

// Config holds the authentication settings.
type Config struct {
	ClientID string
	TenantID string

	ClientSecret string

	// Base64 encoded PKCS#12 bundle, or a path to it
	ClientCertificate         string
	ClientCertificatePath     string
	ClientCertificatePassword string

	UseOIDC bool
	UseMSI  bool

	// OIDC token sources, see FromEnvironment.
	OIDCToken           string
	OIDCTokenFilePath   string
	OIDCRequestURL      string
	OIDCRequestToken    string
	ServiceConnectionID string
}

// MethodConfigured reports whether an authentication method is configured.
func (c Config) MethodConfigured() bool {
	return c.ClientSecret != "" || c.ClientCertificate != "" || c.ClientCertificatePath != "" || c.UseOIDC || c.UseMSI
}

// FromEnvironment fills the OIDC token sources from the environment.
func (c Config) FromEnvironment() Config {
	c.OIDCToken = firstEnv(EnvOIDCToken)
	c.OIDCTokenFilePath = firstEnv(EnvOIDCTokenFilePath)
	c.OIDCRequestURL = firstEnv(EnvOIDCRequestURL)
	c.OIDCRequestToken = firstEnv(EnvOIDCRequestToken)
	c.ServiceConnectionID = firstEnv(EnvServiceConnectionID)
	return c
}

func firstEnv(names []string) string {
	for _, name := range names {
		if v := os.Getenv(name); v != "" {
			return v
		}
	}
	return ""
}

// NewCredential returns the credential for the configured authentication method.
//
// When no method is configured it returns nil, nil: the caller keeps using the
// driver's built-in default credential handling, which preserves the
// behaviour of previous provider versions.
func NewCredential(ctx context.Context, c Config) (azcore.TokenCredential, error) {
	if !c.MethodConfigured() {
		if c.ClientID != "" || c.TenantID != "" {
			return nil, errors.New("client_id and tenant_id require an authentication method: " +
				"client_secret, client_certificate, client_certificate_path, use_oidc or use_msi")
		}
		return nil, nil
	}

	// the tenant is optional for managed identity only
	onlyMSI := c.UseMSI && c.ClientSecret == "" && c.ClientCertificate == "" && c.ClientCertificatePath == "" && !c.UseOIDC
	if !onlyMSI && (c.ClientID == "" || c.TenantID == "") {
		return nil, errors.New("client_id and tenant_id are required for client secret, client certificate and OIDC authentication")
	}

	var certificate []byte
	if c.ClientCertificate != "" {
		var err error
		if certificate, err = base64.StdEncoding.DecodeString(c.ClientCertificate); err != nil {
			return nil, fmt.Errorf("decoding client_certificate: %w", err)
		}
	}

	var token string
	if c.UseOIDC {
		var err error
		if token, err = oidcToken(c); err != nil {
			return nil, err
		}
		if token == "" && (c.OIDCRequestURL == "" || c.OIDCRequestToken == "") {
			return nil, errors.New("use_oidc = true requires an OIDC token (ARM_OIDC_TOKEN or ARM_OIDC_TOKEN_FILE_PATH) " +
				"or an OIDC token endpoint (ARM_OIDC_REQUEST_URL, ACTIONS_ID_TOKEN_REQUEST_URL or SYSTEM_OIDCREQUESTURI, and " +
				"ARM_OIDC_REQUEST_TOKEN, ACTIONS_ID_TOKEN_REQUEST_TOKEN or SYSTEM_ACCESSTOKEN)")
		}
	}

	credential := &Credential{
		credentials: sdkauth.Credentials{
			Environment: *environments.AzurePublic(),
			ClientID:    c.ClientID,
			TenantID:    c.TenantID,

			ClientSecret:              c.ClientSecret,
			ClientCertificateData:     certificate,
			ClientCertificatePath:     c.ClientCertificatePath,
			ClientCertificatePassword: c.ClientCertificatePassword,

			OIDCAssertionToken:             token,
			OIDCTokenRequestURL:            c.OIDCRequestURL,
			OIDCTokenRequestToken:          c.OIDCRequestToken,
			ADOPipelineServiceConnectionID: c.ServiceConnectionID,

			// Order of precedence: client certificate, client secret, OIDC token,
			// Azure DevOps (service connection set), GitHub Actions, managed identity.
			EnableAuthenticatingUsingClientCertificate: true,
			EnableAuthenticatingUsingClientSecret:      true,
			EnableAuthenticationUsingOIDC:              c.UseOIDC,
			EnableAuthenticationUsingADOPipelineOIDC:   c.UseOIDC,
			EnableAuthenticationUsingGitHubOIDC:        c.UseOIDC,
			EnableAuthenticatingUsingManagedIdentity:   c.UseMSI,
		},
		authorizers: map[string]sdkauth.CachingAuthorizer{},
	}

	// validate the configuration; no token is requested yet
	if _, err := credential.authorizer(ctx, "https://database.windows.net/.default"); err != nil {
		return nil, err
	}
	return credential, nil
}

// oidcToken returns the supplied OIDC token, read from the token file if set.
func oidcToken(c Config) (string, error) {
	token := strings.TrimSpace(c.OIDCToken)
	if c.OIDCTokenFilePath == "" {
		return token, nil
	}

	data, err := os.ReadFile(c.OIDCTokenFilePath)
	if err != nil {
		return "", fmt.Errorf("reading OIDC token from file %q: %w", c.OIDCTokenFilePath, err)
	}
	fileToken := strings.TrimSpace(string(data))
	if token != "" && token != fileToken {
		return "", errors.New("ARM_OIDC_TOKEN and the contents of ARM_OIDC_TOKEN_FILE_PATH differ, remove one or ensure they match")
	}
	return fileToken, nil
}

// Credential adapts go-azure-sdk authorizers to azcore.TokenCredential, which
// is used by the SQL driver and the Azure SDK clients. An authorizer is
// created per requested scope and caches its tokens.
type Credential struct {
	credentials sdkauth.Credentials

	mu          sync.Mutex
	authorizers map[string]sdkauth.CachingAuthorizer
}

var _ azcore.TokenCredential = &Credential{}

func (c *Credential) GetToken(ctx context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	if len(options.Scopes) != 1 {
		return azcore.AccessToken{}, fmt.Errorf("expected exactly one scope, got %d", len(options.Scopes))
	}

	authorizer, err := c.authorizer(ctx, options.Scopes[0])
	if err != nil {
		return azcore.AccessToken{}, err
	}

	// the request is not used by these authorizers
	token, err := authorizer.Token(ctx, nil)
	if err != nil {
		return azcore.AccessToken{}, err
	}
	return azcore.AccessToken{Token: token.AccessToken, ExpiresOn: token.Expiry}, nil
}

func (c *Credential) authorizer(ctx context.Context, scope string) (sdkauth.CachingAuthorizer, error) {
	resource := strings.TrimSuffix(strings.TrimSuffix(scope, "/.default"), "/")

	c.mu.Lock()
	defer c.mu.Unlock()

	if authorizer, ok := c.authorizers[resource]; ok {
		return authorizer, nil
	}

	api := environments.NewApiEndpoint(resource, resource, nil).WithResourceIdentifier(resource)
	authorizer, err := sdkauth.NewAuthorizerFromCredentials(ctx, c.credentials, api)
	if err != nil {
		return nil, fmt.Errorf("configuring authentication: %w", err)
	}
	cached, err := sdkauth.NewCachedAuthorizer(authorizer)
	if err != nil {
		return nil, fmt.Errorf("configuring authentication: %w", err)
	}

	c.authorizers[resource] = cached
	return cached, nil
}
