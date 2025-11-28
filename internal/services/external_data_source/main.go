package external_data_source

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ExternalDataSourceDataSourceModel struct {
	Id           types.String `tfsdk:"id"`
	Database     types.String `tfsdk:"database"`
	Name         types.String `tfsdk:"name"`
	Location     types.String `tfsdk:"location"`
	Credential   types.String `tfsdk:"credential"`
	DataSourceId types.Int64  `tfsdk:"data_source_id"`
}

type ExternalDataSourceResourceModel struct {
	Id           types.String `tfsdk:"id"`
	Database     types.String `tfsdk:"database"`
	Name         types.String `tfsdk:"name"`
	Location     types.String `tfsdk:"location"`
	Credential   types.String `tfsdk:"credential"`
	DataSourceId types.Int64  `tfsdk:"data_source_id"`
}
