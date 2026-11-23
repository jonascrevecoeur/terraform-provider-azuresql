package provider

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"terraform-provider-azuresql/internal/auth"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// authConfig builds the authentication settings from the provider
// configuration and the environment.
//
// Every argument falls back to its ARM_* environment variable when not set in
// the provider block. client_id and tenant_id (ARM_CLIENT_ID, ARM_TENANT_ID)
// are only read from the environment once an authentication method is
// configured, as they are commonly set for other providers.
//
// Values that are unknown during plan are rejected instead of falling back to
// the environment, so the identity used never depends on the evaluation order.
func authConfig(config ProviderConfigModel) (auth.Config, error) {
	for _, value := range []attr.Value{
		config.ClientId, config.TenantId, config.ClientSecret,
		config.ClientCertificate, config.ClientCertificatePath, config.ClientCertificatePassword,
		config.UseOIDC, config.UseMSI,
	} {
		if value.IsUnknown() {
			return auth.Config{}, errors.New("the authentication arguments must be known during plan")
		}
	}

	c := auth.Config{
		ClientID:                  config.ClientId.ValueString(),
		TenantID:                  config.TenantId.ValueString(),
		ClientSecret:              stringValue(config.ClientSecret, "ARM_CLIENT_SECRET"),
		ClientCertificate:         stringValue(config.ClientCertificate, "ARM_CLIENT_CERTIFICATE"),
		ClientCertificatePath:     stringValue(config.ClientCertificatePath, "ARM_CLIENT_CERTIFICATE_PATH"),
		ClientCertificatePassword: stringValue(config.ClientCertificatePassword, "ARM_CLIENT_CERTIFICATE_PASSWORD"),
		UseOIDC:                   boolValue(config.UseOIDC, "ARM_USE_OIDC"),
		UseMSI:                    boolValue(config.UseMSI, "ARM_USE_MSI"),
	}

	if c.MethodConfigured() {
		if c.ClientID == "" {
			c.ClientID = os.Getenv("ARM_CLIENT_ID")
		}
		if c.TenantID == "" {
			c.TenantID = os.Getenv("ARM_TENANT_ID")
		}
	}

	return c.FromEnvironment(), nil
}

func stringValue(value types.String, envVar string) string {
	if v := value.ValueString(); v != "" {
		return v
	}
	return os.Getenv(envVar)
}

func boolValue(value types.Bool, envVar string) bool {
	if !value.IsNull() {
		return value.ValueBool()
	}
	b, _ := strconv.ParseBool(strings.TrimSpace(os.Getenv(envVar)))
	return b
}
