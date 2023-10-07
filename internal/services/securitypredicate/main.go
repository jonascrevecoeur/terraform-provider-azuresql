package securitypredicate

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SecurityPredicateResourceModel struct {
	Id               types.String `tfsdk:"id"`
	Database         types.String `tfsdk:"database"`
	SecurityPolicy   types.String `tfsdk:"security_policy"`
	Table            types.String `tfsdk:"table"`
	PredicateId      types.Int64  `tfsdk:"predicate_id"`
	Rule             types.String `tfsdk:"rule"`
	Type             types.String `tfsdk:"type"`
	BlockRestriction types.String `tfsdk:"block_restriction"`
}
