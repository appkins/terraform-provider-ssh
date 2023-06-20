package provider

import (
	"context"
	"os"
	"time"

	"github.com/appkins/terraform-provider-ssh/internal/remote"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ provider.Provider = &SshProvider{}
)

type SshProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (p *SshProvider) Metadata(_ context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ssh"
	resp.Version = p.version
}

func (p *SshProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "The host to connect to. This can be an IP address or a hostname.",
				Required:            true,
			},
			"port": schema.StringAttribute{
				MarkdownDescription: "The port to connect to the remote host on. Defaults to `22`.",
				Optional:            true,
			},
			"user": schema.StringAttribute{
				MarkdownDescription: "The user to connect to the remote host as. Defaults to `root`.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"private_key": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"private_key_path": schema.StringAttribute{
				Optional: true,
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout in seconds",
				Optional:            true,
			},
			"retry_delay": schema.Int64Attribute{
				MarkdownDescription: "Retry delay in seconds",
				Optional:            true,
			},
		},
	}
}

func (p *SshProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {

	var config SshConfig
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := GetSshConfig(&config)

	if host, found := os.LookupEnv("SSH_HOST"); found {
		if cfg.Host == "" {
			cfg.Host = host
		}
	}

	if user, found := os.LookupEnv("SSH_USER"); found {
		if cfg.User == "" {
			cfg.User = user
		}
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 20 * time.Second
	}

	factory := remote.NewFactory(cfg)

	resp.DataSourceData = factory
	resp.ResourceData = factory
}

func (p *SshProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewScriptResource,
	}
}

func (p *SshProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewMetadataDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SshProvider{
			version: version,
		}
	}
}
