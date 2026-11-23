package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func clearAuthEnv(t *testing.T) {
	for _, name := range []string{
		"ARM_CLIENT_ID", "ARM_TENANT_ID", "ARM_CLIENT_SECRET",
		"ARM_CLIENT_CERTIFICATE", "ARM_CLIENT_CERTIFICATE_PATH", "ARM_CLIENT_CERTIFICATE_PASSWORD",
		"ARM_USE_OIDC", "ARM_USE_MSI", "ARM_OIDC_AZURE_SERVICE_CONNECTION_ID",
	} {
		t.Setenv(name, "")
	}
}

func nullConfig() ProviderConfigModel {
	return ProviderConfigModel{
		ClientId:                  types.StringNull(),
		TenantId:                  types.StringNull(),
		ClientSecret:              types.StringNull(),
		ClientCertificate:         types.StringNull(),
		ClientCertificatePath:     types.StringNull(),
		ClientCertificatePassword: types.StringNull(),
		UseOIDC:                   types.BoolNull(),
		UseMSI:                    types.BoolNull(),
	}
}

func TestAuthConfigFromProviderBlock(t *testing.T) {
	clearAuthEnv(t)
	t.Setenv("ARM_OIDC_AZURE_SERVICE_CONNECTION_ID", "sc")

	config := nullConfig()
	config.ClientId = types.StringValue("client")
	config.TenantId = types.StringValue("tenant")
	config.UseOIDC = types.BoolValue(true)

	c, err := authConfig(config)
	if err != nil {
		t.Fatal(err)
	}
	if c.ClientID != "client" || c.TenantID != "tenant" || !c.UseOIDC || c.ServiceConnectionID != "sc" {
		t.Errorf("unexpected config %+v", c)
	}
}

// Every authentication method falls back to the same ARM_* variables.
func TestAuthConfigFromEnvironment(t *testing.T) {
	cases := map[string]map[string]string{
		"client secret":           {"ARM_CLIENT_SECRET": "secret"},
		"client certificate":      {"ARM_CLIENT_CERTIFICATE": "Y2VydA=="},
		"client certificate path": {"ARM_CLIENT_CERTIFICATE_PATH": "/cert.pfx", "ARM_CLIENT_CERTIFICATE_PASSWORD": "pw"},
		"oidc":                    {"ARM_USE_OIDC": "true"},
		"managed identity":        {"ARM_USE_MSI": "true"},
	}

	for name, env := range cases {
		t.Run(name, func(t *testing.T) {
			clearAuthEnv(t)
			t.Setenv("ARM_CLIENT_ID", "env-client")
			t.Setenv("ARM_TENANT_ID", "env-tenant")
			for k, v := range env {
				t.Setenv(k, v)
			}

			c, err := authConfig(nullConfig())
			if err != nil {
				t.Fatal(err)
			}
			if !c.MethodConfigured() || c.ClientID != "env-client" || c.TenantID != "env-tenant" {
				t.Errorf("unexpected config %+v", c)
			}
			if env["ARM_CLIENT_CERTIFICATE_PASSWORD"] != "" && c.ClientCertificatePassword != "pw" {
				t.Errorf("certificate password not read, got %+v", c)
			}
		})
	}
}

func TestAuthConfigProviderBlockTakesPrecedence(t *testing.T) {
	clearAuthEnv(t)
	t.Setenv("ARM_CLIENT_ID", "env-client")
	t.Setenv("ARM_TENANT_ID", "env-tenant")
	t.Setenv("ARM_CLIENT_SECRET", "env-secret")

	config := nullConfig()
	config.ClientId = types.StringValue("client")
	config.TenantId = types.StringValue("tenant")
	config.ClientSecret = types.StringValue("secret")

	c, err := authConfig(config)
	if err != nil {
		t.Fatal(err)
	}
	if c.ClientID != "client" || c.TenantID != "tenant" || c.ClientSecret != "secret" {
		t.Errorf("unexpected config %+v", c)
	}
}

// use_oidc = false / use_msi = false in the provider block opt out of ARM_USE_OIDC / ARM_USE_MSI.
func TestAuthConfigOptOut(t *testing.T) {
	clearAuthEnv(t)
	t.Setenv("ARM_USE_OIDC", "true")
	t.Setenv("ARM_USE_MSI", "true")
	t.Setenv("ARM_CLIENT_ID", "env-client")

	config := nullConfig()
	config.UseOIDC = types.BoolValue(false)
	config.UseMSI = types.BoolValue(false)

	c, err := authConfig(config)
	if err != nil {
		t.Fatal(err)
	}
	if c.MethodConfigured() || c.ClientID != "" {
		t.Errorf("unexpected config %+v", c)
	}
}

// Without an authentication method, ARM_CLIENT_ID and ARM_TENANT_ID are not picked up
// and the default credential chain is used.
func TestAuthConfigIdentityIgnoredWithoutMethod(t *testing.T) {
	clearAuthEnv(t)
	t.Setenv("ARM_CLIENT_ID", "env-client")
	t.Setenv("ARM_TENANT_ID", "env-tenant")

	c, err := authConfig(nullConfig())
	if err != nil {
		t.Fatal(err)
	}
	if c.MethodConfigured() || c.ClientID != "" || c.TenantID != "" {
		t.Errorf("unexpected config %+v", c)
	}
}

func TestAuthConfigInvalidBoolIgnored(t *testing.T) {
	clearAuthEnv(t)
	t.Setenv("ARM_USE_OIDC", "maybe")

	c, err := authConfig(nullConfig())
	if err != nil {
		t.Fatal(err)
	}
	if c.UseOIDC {
		t.Error("an unparsable ARM_USE_OIDC must not enable OIDC")
	}
}

func TestAuthConfigUnknownValues(t *testing.T) {
	clearAuthEnv(t)
	t.Setenv("ARM_CLIENT_SECRET", "env-secret")

	config := nullConfig()
	config.ClientSecret = types.StringUnknown()

	if _, err := authConfig(config); err == nil {
		t.Error("expected an error for unknown values")
	}
}
