package synapseserver

import "github.com/hashicorp/terraform-plugin-framework/types"

type synapseserverDataSourceModel struct {
	ConnectionId types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Port         types.Int64  `tfsdk:"port"`
	Serverless   types.Bool   `tfsdk:"serverless"`
}
