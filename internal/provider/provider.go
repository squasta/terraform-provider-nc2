package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nutanix/terraform-provider-nc2/internal/client"
)

// version is overridden via -ldflags at release time.
var version = "dev" //nolint:gochecknoglobals

// Per-story registry slices populated by the registry_*.go init
// blocks. Kept package-level so each US-phase init can append to
// them without depending on the others.
//
//nolint:gochecknoglobals // documented per-story registration mechanism.
var (
	resourceRegistry   []func() resource.Resource
	dataSourceRegistry []func() datasource.DataSource
	actionRegistry     []func() action.Action
)

// Runtime is the bundle the provider hands to every resource /
// data source / action via the framework's ProviderData field.
// Implements both orgshared.RuntimeAccessor and
// dsshared.RuntimeAccessor through a single NC2Client method.
type Runtime struct {
	client *client.Client
}

// NC2Client returns the configured NC2 HTTP client. May be nil if
// Configure has not yet completed successfully.
func (r *Runtime) NC2Client() *client.Client { return r.client }

// New constructs the Provider for the framework. Wired by main.go
// via providerserver.Serve.
func New() provider.Provider { return &nc2Provider{} }

type nc2Provider struct{}

// Metadata returns the provider's type name and version.
func (p *nc2Provider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "nc2"
	resp.Version = version
}

// Schema returns the provider configuration schema. Mirrors
// contracts/provider.json.
func (p *nc2Provider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Nutanix Cloud Clusters (NC2) provider. See https://nutanix.dev for the underlying API.",
		Attributes: map[string]schema.Attribute{
			"api_key":          schema.StringAttribute{Optional: true, Sensitive: true, Description: "MyNutanix API key (FR-001)."},
			"key_id":           schema.StringAttribute{Optional: true, Sensitive: true, Description: "MyNutanix API key id; JWT kid header."},
			"issuer":           schema.StringAttribute{Optional: true, Sensitive: true, Description: "JWT iss claim (organization UUID)."},
			"credentials_file": schema.StringAttribute{Optional: true, Description: "Path to a credentials file (defaults to ~/.nc2/credentials)."},
			"profile":          schema.StringAttribute{Optional: true, Description: "Profile name within the credentials file."},
			"base_url":         schema.StringAttribute{Optional: true, Description: "Override of the NC2 API base URL."},
			"ca_bundle":        schema.StringAttribute{Optional: true, Description: "Path to a PEM file appended to the system trust store (FR-003b)."},
			"task_poll_interval_seconds": schema.Int64Attribute{Optional: true, Description: "Interval between NC2 task polls (FR-003)."},
			"task_max_timeout_seconds":   schema.Int64Attribute{Optional: true, Description: "Maximum total time spent polling a single NC2 task."},
		},
	}
}

// providerModel mirrors the framework Schema above for tfsdk
// decoding inside Configure.
type providerModel struct {
	APIKey                  types.String `tfsdk:"api_key"`
	KeyID                   types.String `tfsdk:"key_id"`
	Issuer                  types.String `tfsdk:"issuer"`
	CredentialsFile         types.String `tfsdk:"credentials_file"`
	Profile                 types.String `tfsdk:"profile"`
	BaseURL                 types.String `tfsdk:"base_url"`
	CABundle                types.String `tfsdk:"ca_bundle"`
	TaskPollIntervalSeconds types.Int64  `tfsdk:"task_poll_interval_seconds"`
	TaskMaxTimeoutSeconds   types.Int64  `tfsdk:"task_max_timeout_seconds"`
}

// Configure builds the runtime bundle from the provider block, env,
// and credentials_file, and hands it to every resource / data source
// / action via ProviderData.
func (p *nc2Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	block := ProviderBlock{
		APIKey:          cfg.APIKey.ValueString(),
		KeyID:           cfg.KeyID.ValueString(),
		Issuer:          cfg.Issuer.ValueString(),
		CredentialsFile: cfg.CredentialsFile.ValueString(),
		Profile:         cfg.Profile.ValueString(),
		BaseURL:         cfg.BaseURL.ValueString(),
		CABundle:        cfg.CABundle.ValueString(),
		TaskPollSeconds: cfg.TaskPollIntervalSeconds.ValueInt64(),
		TaskMaxSeconds:  cfg.TaskMaxTimeoutSeconds.ValueInt64(),
	}
	env := EnvFromOS()
	file, err := ReadFile(firstNonEmpty(block.CredentialsFile, env.CredentialsFile))
	if err != nil {
		resp.Diagnostics.AddError("nc2 provider: credentials_file unreadable", err.Error())
		return
	}

	pcfg, diags := Resolve(ctx, block, env, file)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tlsCfg, err := BuildTLSConfig(pcfg.CABundle)
	if err != nil {
		resp.Diagnostics.AddError("nc2 provider: ca_bundle invalid", err.Error())
		return
	}

	cli, err := client.New(client.Config{
		Credentials:      pcfg.Credentials,
		BaseURL:          pcfg.BaseURL,
		UserAgent:        "terraform-provider-nc2/" + version,
		TLSConfig:        tlsCfg,
		TaskPollInterval: time.Duration(pcfg.TaskPollSeconds) * time.Second,
		TaskMaxTimeout:   time.Duration(pcfg.TaskMaxSeconds) * time.Second,
	})
	if err != nil {
		resp.Diagnostics.AddError("nc2 provider: building HTTP client", err.Error())
		return
	}

	rt := &Runtime{client: cli}
	resp.ResourceData = rt
	resp.DataSourceData = rt
	resp.EphemeralResourceData = rt
	resp.ActionData = rt
}

// Resources returns every constructor registered via
// registry_*.go init functions.
func (p *nc2Provider) Resources(_ context.Context) []func() resource.Resource {
	out := make([]func() resource.Resource, len(resourceRegistry))
	copy(out, resourceRegistry)
	return out
}

// DataSources returns every constructor registered via
// registry_*.go init functions.
func (p *nc2Provider) DataSources(_ context.Context) []func() datasource.DataSource {
	out := make([]func() datasource.DataSource, len(dataSourceRegistry))
	copy(out, dataSourceRegistry)
	return out
}

// Actions returns every constructor registered via registry_*.go
// init functions.
func (p *nc2Provider) Actions(_ context.Context) []func() action.Action {
	out := make([]func() action.Action, len(actionRegistry))
	copy(out, actionRegistry)
	return out
}

var _ provider.Provider = &nc2Provider{}
var _ provider.ProviderWithActions = &nc2Provider{}
