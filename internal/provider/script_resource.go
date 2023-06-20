package provider

import (
	"context"
	"fmt"

	"github.com/appkins/terraform-provider-ssh/internal/remote"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	Destroy = "destroy"
	Create  = "create"
	Update  = "update"
	Read    = "read"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ScriptResource{}
var _ resource.ResourceWithImportState = &ScriptResource{}

func NewScriptResource() resource.Resource {
	return &ScriptResource{}
}

// ScriptResource defines the resource implementation.
type ScriptResource struct {
	factory *remote.ProvisionerFactory
}

// ScriptResourceModel describes the resource data model.
type ScriptResourceModel struct {
	Triggers   types.Map    `tfsdk:"triggers"`
	Timeout    types.String `tfsdk:"timeout"`
	RetryDelay types.String `tfsdk:"retry_delay"`
	//Connect    types.Set    `tfsdk:"connect"`
	//Query      types.Set    `tfsdk:"query"`
	//Script     types.Set    `tfsdk:"script"`
	Exec []struct {
		Commands  []types.String `tfsdk:"commands"`
		Lifecycle types.String   `tfsdk:"lifecycle"`
	} `tfsdk:"exec"`
	File []struct {
		Source      types.String `tfsdk:"source"`
		Destination types.String `tfsdk:"destination"`
		Content     types.String `tfsdk:"content"`
		Permissions types.String `tfsdk:"permissions"`
		Owner       types.String `tfsdk:"owner"`
		Group       types.String `tfsdk:"group"`
	} `tfsdk:"file"`
	Ssh *SshConfig `tfsdk:"ssh"`

	Result struct {
		Read    []types.String `tfsdk:"read"`
		Update  []types.String `tfsdk:"update"`
		Create  []types.String `tfsdk:"create"`
		Destroy []types.String `tfsdk:"destroy"`
	} `tfsdk:"result"`
}

func (r *ScriptResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_script"
}

func (r *ScriptResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Script resource",

		Attributes: map[string]schema.Attribute{
			"triggers": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "A map of arbitrary strings that, when changed, will force the 'hsdp_container_host_exec' resource to be replaced, re-running any associated commands.",
				Optional:            true,
			},
			"timeout": schema.StringAttribute{
				MarkdownDescription: "Timeout for the SSH connection.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("5m"),
			},
			"retry_delay": schema.StringAttribute{
				MarkdownDescription: "Delay before retrying the SSH connection.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("10s"),
			},
			"result": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					Read: schema.ListAttribute{
						ElementType: types.StringType,
						Computed:    true,
					},
					Update: schema.ListAttribute{
						ElementType: types.StringType,
						Computed:    true,
					},
					Create: schema.ListAttribute{
						ElementType: types.StringType,
						Computed:    true,
					},
					Destroy: schema.ListAttribute{
						ElementType: types.StringType,
						Computed:    true,
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"file": schema.SetNestedBlock{
				MarkdownDescription: "Files.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"source": schema.StringAttribute{
							MarkdownDescription: "Source path to the file to be copied.",
							Optional:            true,
						},
						"content": schema.StringAttribute{
							Optional:  true,
							Sensitive: true,
						},
						"destination": schema.StringAttribute{
							Required: true,
						},
						"permissions": schema.StringAttribute{
							Optional: true,
						},
						"owner": schema.StringAttribute{
							Optional: true,
						},
						"group": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			"exec": schema.SetNestedBlock{
				MarkdownDescription: "Commands to execute.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"commands": schema.SetAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "List of commands to run.",
							Required:            true,
						},
						"lifecycle": schema.StringAttribute{
							MarkdownDescription: "Lifecycle of the command. Valid values are `create`, `read`, `update` and `destroy`.",
							Optional:            true,
						},
					},
				},
			},
			"ssh": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"host": schema.StringAttribute{
						MarkdownDescription: "SSH host",
						Optional:            true,
					},
					"port": schema.StringAttribute{
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

func (r *ScriptResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
		r.factory = factory
	}
}

func (r *ScriptResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ScriptResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.factory.Create(GetSshConfig(data.Ssh))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create SSH client",
			fmt.Sprintf("Failed to create SSH client: %s", err.Error()),
		)
		return
	}

	var scripts []string
	var files []remote.File

	for _, e := range data.Exec {
		if e.Lifecycle.String() == Create {
			for _, c := range e.Commands {
				scripts = append(scripts, c.String())
			}
		}
	}

	for _, f := range data.File {
		files = append(files, remote.File{
			Source:      f.Source,
			Destination: f.Destination,
			Content:     f.Content,
			Permissions: f.Permissions,
			Owner:       f.Owner,
			Group:       f.Group,
		})
	}

	if err := client.CopyFiles(ctx, files); err != nil {
		resp.Diagnostics.AddError(
			"Failed to copy files",
			fmt.Sprintf("Failed to copy files: %s", err.Error()),
		)
		return
	}

	if out, err := client.Execute(ctx, scripts); err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute script",
			fmt.Sprintf("Failed to execute script: %s", err.Error()),
		)
	} else {
		var outVals []types.String
		for _, o := range out {
			outVals = append(outVals, types.StringValue(o))
		}
		data.Result.Create = outVals
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScriptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ScriptResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	scripts := make([]string, 0)

	for _, e := range data.Exec {
		if e.Lifecycle.ValueString() == "read" {
			for _, c := range e.Commands {
				scripts = append(scripts, c.String())
			}
		}
	}

	client, err := r.factory.Create(GetSshConfig(data.Ssh))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create SSH client",
			fmt.Sprintf("Failed to create SSH client: %s", err.Error()),
		)
		return
	}

	if out, err := client.Execute(ctx, scripts); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read script, got error: %s", err))
	} else {
		var outVals []types.String
		for _, o := range out {
			outVals = append(outVals, types.StringValue(o))
		}
		data.Result.Read = outVals
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScriptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ScriptResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.factory.Create(GetSshConfig(data.Ssh))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create SSH client",
			fmt.Sprintf("Failed to create SSH client: %s", err.Error()),
		)
		return
	}

	scripts := make([]string, 0)
	for _, e := range data.Exec {
		if e.Lifecycle.ValueString() == Update {
			for _, c := range e.Commands {
				scripts = append(scripts, c.String())
			}
		}
	}

	if out, err := client.Execute(ctx, scripts); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read script, got error: %s", err))
	} else {
		var outVals []types.String
		for _, o := range out {
			outVals = append(outVals, types.StringValue(o))
		}
		data.Result.Update = outVals
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScriptResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ScriptResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.factory.Create(GetSshConfig(data.Ssh))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create SSH client",
			fmt.Sprintf("Failed to create SSH client: %s", err.Error()),
		)
		return
	}

	scripts := make([]string, 0)
	for _, e := range data.Exec {
		if e.Lifecycle.String() == Destroy {
			for _, c := range e.Commands {
				scripts = append(scripts, c.String())
			}
		}
	}

	if out, err := client.Execute(ctx, scripts); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete script, got error: %s", err))
	} else {
		var outVals []types.String
		for _, o := range out {
			outVals = append(outVals, types.StringValue(o))
		}
		data.Result.Create = outVals
	}
}

func (r *ScriptResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
