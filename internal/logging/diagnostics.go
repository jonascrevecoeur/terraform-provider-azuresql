package logging

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

const azuresqlProviderLoggingDiagnosticsKey = "terraform.azuresql.logging.diagnostics"

func GetTestContext() (ctx context.Context) {
	return WithDiagnostics(context.Background(), &diag.Diagnostics{})
}

func GetDiagnostics(ctx context.Context) *diag.Diagnostics {
	return ctx.Value(azuresqlProviderLoggingDiagnosticsKey).(*diag.Diagnostics)
}

func WithDiagnostics(ctx context.Context, diagnostics *diag.Diagnostics) context.Context {
	return context.WithValue(ctx, azuresqlProviderLoggingDiagnosticsKey, diagnostics)
}

func HasError(ctx context.Context) bool {
	return GetDiagnostics(ctx).HasError()
}

func AddError(ctx context.Context, summary string, err interface{}) {
	if err == nil {
		return
	}

	switch v := err.(type) {
	case string:
		GetDiagnostics(ctx).AddError(summary, err.(string))
	case error:
		GetDiagnostics(ctx).AddError(summary, err.(error).Error())
	default:
		GetDiagnostics(ctx).AddError("Invalid type for err in logging.AddError",
			fmt.Sprintf("Object of type %s provided. Only types string and error are supported.", v))
	}
}

func AddWarning(ctx context.Context, summary string, err interface{}) {
	if err == nil {
		return
	}

	switch v := err.(type) {
	case string:
		GetDiagnostics(ctx).AddWarning(summary, err.(string))
	case error:
		GetDiagnostics(ctx).AddWarning(summary, err.(error).Error())
	default:
		GetDiagnostics(ctx).AddWarning("Invalid type for err in logging.AddWarning",
			fmt.Sprintf("Object of type %s provided. Only types string and error are supported.", v))
	}
}

func AddAttributeError(ctx context.Context, path path.Path, summary string, details string) {
	GetDiagnostics(ctx).AddAttributeError(path, summary, details)
}

func AppendDiagnostics(ctx context.Context, diagnostics ...diag.Diagnostic) {
	GetDiagnostics(ctx).Append(diagnostics...)
}

func ClearDiagnostics(ctx context.Context) {
	*GetDiagnostics(ctx) = nil
}
