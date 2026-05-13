// Package gcp_cluster implements the nc2_gcp_cluster managed resource.
// Mirrors AWS minus AWS-only fields. GCP-specific subfields live
// inside the shared `network` Map<String,String>.
package gcp_cluster //nolint:revive,staticcheck

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/nutanix/terraform-provider-nc2/internal/resources/clustershared"
)

type model struct {
	clustershared.Model
}

func resourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages an NC2 cluster on Google Cloud Platform. " +
			"`access_policy` is intentionally absent (AWS-only per FR-010a).",
		Attributes: clustershared.CommonAttributes(),
	}
}
