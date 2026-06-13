package provider

import (
	"context"
	"fmt"

	"github.com/appkins/terraform-provider-ssh/internal/ssh"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &FileDataSource{}

func NewFileDataSource() datasource.DataSource {
	return &FileDataSource{}
}

// FileDataSource defines the data source implementation.
type FileDataSource struct {
	factory *ssh.Factory
}

// FileDataSourceModel describes the data source data model.
type FileDataSourceModel struct {
	Ssh     *ssh.Config  `tfsdk:"ssh"`
	Path    types.String `tfsdk:"path"`
	Content types.String `tfsdk:"content"`
}

func (d *FileDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file"
}

func (d *FileDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "File data source",

		Attributes: map[string]schema.Attribute{
			"path": schema.StringAttribute{
				MarkdownDescription: "File identifier",
				Required:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "File data.",
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"ssh": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"host": schema.StringAttribute{
						MarkdownDescription: "SSH host",
						Optional:            true,
					},
					"port": schema.Int64Attribute{
						MarkdownDescription: "SSH port",
						Optional:            true,
					},
					"user": schema.StringAttribute{
						MarkdownDescription: "SSH user",
						Optional:            true,
					},
					"password": schema.StringAttribute{
						MarkdownDescription: "SSH password",
						Optional:            true,
						Sensitive:           true,
					},
					"private_key": schema.StringAttribute{
						MarkdownDescription: "SSH private key data",
						Optional:            true,
						Sensitive:           true,
					},
					"private_key_path": schema.StringAttribute{
						MarkdownDescription: "Path to SSH private key",
						Optional:            true,
					},
					"timeout": schema.Int64Attribute{
						MarkdownDescription: "Timeout",
						Optional:            true,
					},
					"retry_delay": schema.Int64Attribute{
						MarkdownDescription: "Retry delay",
						Optional:            true,
					},
				},
			},
		},
	}
}

func (d *FileDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	if factory, ok := req.ProviderData.(*ssh.Factory); !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	} else {
		d.factory = factory
	}
}

func (d *FileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FileDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.factory.Create(*data.Ssh)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create SSH client",
			fmt.Sprintf("Failed to create SSH client: %s", err.Error()),
		)
		return
	}

	path := data.Path.ValueString()

	if stdOut, err := client.Exec(fmt.Sprintf("cat %s", path)); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read file, got error: %s", err))
		return
	} else if len(stdOut) > 0 {
		data.Content = types.StringValue(stdOut)
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read file, got error: %s", err))
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
