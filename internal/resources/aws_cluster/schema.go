// Package aws_cluster implements the nc2_aws_cluster managed resource.
//
// It composes the shared cluster surface (clustershared.CommonAttributes)
// with one AWS-only attribute, `access_policy`, that maps to
// /clusters/{id}/update-access-policy. Setting `access_policy` on
// other clouds is rejected at compile time (the attribute does not
// exist there) per FR-010a.
package aws_cluster //nolint:revive,staticcheck

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nutanix/terraform-provider-nc2/internal/resources/clustershared"
)

// model is the AWS-cluster framework model. It embeds the shared
// fields and adds the AWS-only `access_policy` map.
type model struct {
	clustershared.Model
	AccessPolicy types.Map `tfsdk:"access_policy"`
}

func resourceSchema(_ context.Context) schema.Schema {
	attrs := clustershared.CommonAttributes()
	attrs["access_policy"] = schema.MapAttribute{
		Description: "AWS-only access policy. Setting this attribute on nc2_azure_cluster or " +
			"nc2_gcp_cluster is rejected at the schema level (FR-010a). " +
			"In-place changes route to POST /clusters/{id}/update-access-policy.",
		ElementType: types.StringType,
		Optional:    true,
	}
	return schema.Schema{
		Description: "Manages an NC2 cluster on AWS. Async create/update/delete are handled via internal task polling (FR-008).",
		Attributes:  attrs,
	}
}
