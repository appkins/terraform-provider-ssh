package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/appkins/terraform-provider-ssh/internal/remote"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &MetadataDataSource{}

var (
	cmds = []string{
		"hostname",
		"ip a | grep global | grep -v '10.0.2.15' | awk '{print $2}' | cut -f1 -d '/' | paste -s -d, -",
	}
)

func NewMetadataDataSource() datasource.DataSource {
	return &MetadataDataSource{}
}

// MetadataDataSource defines the data source implementation.
type MetadataDataSource struct {
	client *remote.Provisioner
}

// MetadataDataSourceModel describes the data source data model.
type MetadataDataSourceModel struct {
	Ssh         SshConfig      `tfsdk:"configurable_attribute"`
	HostName    types.String   `tfsdk:"hostname"`
	IpAddress   types.String   `tfsdk:"ip_address"`
	IpAddresses []types.String `tfsdk:"ip_addresses"`
}

func (d *MetadataDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metadata"
}

func (d *MetadataDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Metadata data source",

		Attributes: map[string]schema.Attribute{
			"ssh": schema.StringAttribute{
				MarkdownDescription: "Metadata configurable attribute",
				Optional:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Metadata identifier",
				Computed:            true,
			},
			"ip_address": schema.StringAttribute{
				MarkdownDescription: "Primary private address of the host.",
				Computed:            true,
			},
			"ip_addresses": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "All available private addresses of the host.",
				Computed:            true,
			},
		},
		Blocks: SshDatasourceBlock,
	}
}

func (d *MetadataDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	if client, ok := req.ProviderData.(*remote.Provisioner); !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	} else {
		d.client = client
	}
}

func (d *MetadataDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MetadataDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	stdOut, err := d.client.Execute(ctx, cmds, GetSshConfig(data.Ssh))

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read metadata, got error: %s", err))
		return
	}

	for i, line := range strings.Split(stdOut, "\n") {
		switch i {
		case 0:
			data.HostName = types.StringValue(line)
		case 1:
			{
				if len(line) == 0 {
					break
				}
				if strings.Contains(line, ",") {
					ipaddresses := make([]types.String, 0)
					for i, ip := range strings.Split(line, ",") {
						if i == 0 {
							data.IpAddress = types.StringValue(ip)
						}
						ipaddresses = append(ipaddresses, types.StringValue(ip))
					}
					data.IpAddresses = ipaddresses
					break
				}
			}
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
