package user_default_schema

import "github.com/hashicorp/terraform-plugin-framework/types"

type UserDefaultSchemaResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Database      types.String `tfsdk:"database"`
	User          types.String `tfsdk:"user"`
	DefaultSchema types.String `tfsdk:"default_schema"`
}
