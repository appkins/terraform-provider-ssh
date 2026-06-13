package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/appkins/terraform-provider-ssh/internal/remote"
	"github.com/appkins/terraform-provider-ssh/internal/ssh"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &MetadataDataSource{}

func NewMetadataDataSource() datasource.DataSource {
	return &MetadataDataSource{}
}

// MetadataDataSource defines the data source implementation.
type MetadataDataSource struct {
	factory *remote.ProvisionerFactory
}

type MetadataOs struct {
	Name    types.String `tfsdk:"name"`
	Version types.String `tfsdk:"version"`
	Family  types.String `tfsdk:"family"`
}

// MetadataDataSourceModel describes the data source data model.
type MetadataDataSourceModel struct {
	Ssh         *ssh.Config    `tfsdk:"ssh"`
	HostName    types.String   `tfsdk:"hostname"`
	IpAddress   types.String   `tfsdk:"ip_address"`
	IpAddresses []types.String `tfsdk:"ip_addresses"`
	Os          *MetadataOs    `tfsdk:"os"`
	Raw         types.String   `tfsdk:"raw"`
}

func (d *MetadataDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metadata"
}

func (d *MetadataDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Metadata data source",

		Attributes: map[string]schema.Attribute{
			"raw": schema.StringAttribute{
				MarkdownDescription: "Raw output from metadata commands.",
				Computed:            true,
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
			"os": schema.SingleNestedAttribute{
				MarkdownDescription: "Operating system metadata",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "Operating system name",
						Optional:            true,
						Computed:            true,
					},
					"version": schema.StringAttribute{
						MarkdownDescription: "Operating system version",
						Optional:            true,
						Computed:            true,
					},
					"family": schema.StringAttribute{
						MarkdownDescription: "Operating system family",
						Optional:            true,
						Computed:            true,
					},
				},
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

func (d *MetadataDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	if factory, ok := req.ProviderData.(*remote.ProvisionerFactory); !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	} else {
		d.factory = factory
	}
}

func (d *MetadataDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MetadataDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.factory.Create(ctx, data.Ssh)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create SSH client",
			fmt.Sprintf("Failed to create SSH client: %s", err.Error()),
		)
		return
	}

	cmds := []string{
		"hostname",
		"ip a | grep global | grep -v '10.0.2.15' | awk '{print $2}' | cut -f1 -d '/' | paste -s -d, -",
		"lsb_release -d | awk '{print $2} {print $3} {print $4}'",
	}

	if stdOut, err := client.Execute(cmds); err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute command",
			fmt.Sprintf("Failed to execute command: %s", err.Error()),
		)
		return
	} else {
		rawb := new(strings.Builder)
		var sani []string
		if stdOut[len(stdOut)-1] == "" {
			sani = stdOut[:len(stdOut)-1]
		} else {
			sani = stdOut
		}
		for i, line := range sani {
			rawb.WriteString(line)
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
			case 2:
				{
					if len(line) == 0 {
						break
					}
					if strings.Contains(line, "\n") {
						if data.Os == nil {
							data.Os = new(MetadataOs)
						}
						for ii, l := range strings.Split(line, "\n") {
							switch ii {
							case 0:
								data.Os.Name = types.StringValue(l)
							case 1:
								data.Os.Version = types.StringValue(l)
							case 2:
								data.Os.Family = types.StringValue(l)
							}
						}
					}
				}
			}
		}
		data.Raw = types.StringValue(rawb.String())
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
