package provider

import (
	"context"
	"os"
	"time"

	"github.com/appkins/terraform-provider-ssh/internal/remote"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/loafoe/easyssh-proxy/v2"
)

var (
	_ provider.Provider = &frameworkProvider{}
)

func New() provider.Provider {
	return &frameworkProvider{}
}

type frameworkProvider struct{}

func (p *frameworkProvider) Metadata(_ context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ssh"
}

func (p *frameworkProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Required: true,
			},
			"port": schema.StringAttribute{
				Optional: true,
			},
			"user": schema.StringAttribute{
				Optional: true,
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
		},
	}
}

func (p *frameworkProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {

	var config SshConfig
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown HashiCups API Host",
			"The provider cannot create the HashiCups API client as there is an unknown configuration value for the HashiCups API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_HOST environment variable.",
		)
	}

	if config.User.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown HashiCups API User",
			"The provider cannot create the HashiCups API client as there is an unknown configuration value for the HashiCups API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() && config.PrivateKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown HashiCups API Password",
			"The provider cannot create the HashiCups API client as there is an unknown configuration value for the HashiCups API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	host := os.Getenv("SSH_HOST")
	user := os.Getenv("SSH_USER")
	password := os.Getenv("SSH_PASSWORD")
	private_key := os.Getenv("SSH_PRIVATE_KEY")
	private_key_path := os.Getenv("SSH_PRIVATE_KEY_PATH")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.User.IsNull() {
		user = config.User.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	if !config.PrivateKey.IsNull() {
		private_key = config.PrivateKey.ValueString()
	}

	if !config.PrivateKeyPath.IsNull() {
		private_key_path = config.PrivateKeyPath.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing SSH Host",
			"The provider cannot create the SSH client as there is a missing or empty value for the SSH host. "+
				"Set the host value in the configuration or use the SSH_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if user == "" {
		user = "root"
	}

	if password == "" && private_key == "" && private_key_path == "" {
		resp.Diagnostics.AddError(
			"Missing Credentials",
			"The provider cannot create the SSH client as no value is provided for password, private_key or private_key_path. "+
				"Set the password value in the configuration or use the SSH_PASSWORD environment variable. "+
				"Set the private_key value in the configuration or use the SSH_PASSWORD environment variable. "+
				"Set the private_key_path value in the configuration or use the SSH_PASSWORD environment variable. "+
				"If a value is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	t1, _ := time.ParseDuration("20s")

	client := remote.NewProvisioner(&easyssh.MakeConfig{
		User:     user,
		Server:   host,
		Password: password,
		Key:      private_key,
		KeyPath:  private_key_path,
	}, t1, t1)

	//client := operator.NewSSHOperator()

	// client.Ssh.KeyPath = "~/.ssh/id_ed25519"

	// Create a new HashiCups client using the configuration values
	// client, err := client.AddClient(cfg.Host.String(), &ssh.ClientConfig{User: user, Auth: []ssh.AuthMethod{ssh.Password(password)}})
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Unable to Create HashiCups API Client",
	// 		"An unexpected error occurred when creating the HashiCups API client. "+
	// 			"If the error is not clear, please contact the provider developers.\n\n"+
	// 			"HashiCups Client Error: "+err.Error(),
	// 	)
	// 	return
	// }

	// Make the HashiCups client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *frameworkProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewScriptResource,
	}
}

func (p *frameworkProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		//NewExampleDataSource,
	}
}
