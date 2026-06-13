package provider

import (
	"context"
	"fmt"

	"github.com/appkins/terraform-provider-ssh/internal/ssh"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &FileResource{}
var _ resource.ResourceWithImportState = &FileResource{}

func NewFileResource() resource.Resource {
	return &FileResource{}
}

// FileResource defines the resource implementation.
type FileResource struct {
	factory *ssh.Factory
}

// FileResourceModel describes the resource data model.
type FileResourceModel struct {
	Source      types.String `tfsdk:"source"`
	Destination types.String `tfsdk:"destination"`
	Content     types.String `tfsdk:"content"`
	Permissions types.String `tfsdk:"permissions"`
	Owner       types.String `tfsdk:"owner"`
	Group       types.String `tfsdk:"group"`

	TriggersReplace types.List  `tfsdk:"triggers_replace"`
	Ssh             *ssh.Config `tfsdk:"ssh"`

	Result types.String `tfsdk:"result"`
}

func (r *FileResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file"
}

func (r *FileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This defileion is used by the documentation generator and the language server.
		MarkdownDescription: "File resource",

		Attributes: map[string]schema.Attribute{
			"source": schema.StringAttribute{
				MarkdownDescription: "Source path to the file to be copied.",
				Optional:            true,
			},
			"content": schema.StringAttribute{
				Optional: true,
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
			"triggers_replace": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "A value which is stored in the instance state, and will force replacement when the value changes.",
				Optional:            true,
			},
			"result": schema.StringAttribute{
				Computed: true,
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
						Computed:            true,
						Default:             int64default.StaticInt64(22),
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
						Computed:            true,
						Default:             int64default.StaticInt64(15),
					},
					"retry_delay": schema.Int64Attribute{
						MarkdownDescription: "Retry delay",
						Optional:            true,
						Computed:            true,
						Default:             int64default.StaticInt64(15),
					},
				},
			},
		},
	}
}

func (r *FileResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	if factory, ok := req.ProviderData.(*ssh.Factory); !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	} else {
		r.factory = factory
	}
}

func (r *FileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *FileResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.factory.Create(*data.Ssh)

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create SSH client",
			fmt.Sprintf("Failed to create SSH client: %s", err.Error()),
		)
		return
	}

	resp.Diagnostics.AddWarning("SSH client created", fmt.Sprintf("SSH client created: %s", data.Destination.ValueString()))

	if err := client.WriteFileToHost(ssh.File{
		Source:      data.Source,
		Destination: data.Destination,
		Content:     data.Content,
		Permissions: data.Permissions,
		Owner:       data.Owner,
		Group:       data.Group,
	}); err != nil {
		resp.Diagnostics.AddError(
			"Failed to copy files",
			fmt.Sprintf("Failed to copy files: %s", err.Error()),
		)
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *FileResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.factory.Create(*data.Ssh)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create SSH client",
			fmt.Sprintf("Failed to create SSH client: %s", err.Error()),
		)
		return
	}

	if data.Result.ValueString() != "" {

		if out, err := client.Exec(fmt.Sprintf("cat %s", data.Destination.ValueString())); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read file, got error: %s", err))
		} else {
			data.Result = types.StringValue(out)
		}

		// Save updated data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	}
}

func (r *FileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *FileResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, _ := r.factory.Create(*data.Ssh)

	if err := client.WriteFileToHost(ssh.File{
		Source:      data.Source,
		Destination: data.Destination,
		Content:     data.Content,
		Permissions: data.Permissions,
		Owner:       data.Owner,
		Group:       data.Group,
	}); err != nil {
		resp.Diagnostics.AddError(
			"Failed to copy files",
			fmt.Sprintf("Failed to copy files: %s", err.Error()),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *FileResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// client, err := r.factory.Create(*data.Ssh)

	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Failed to create SSH client",
	// 		fmt.Sprintf("Failed to create SSH client: %s", err.Error()),
	// 	)
	// 	return
	// }

	//if _, err := client.Exec(fmt.Sprintf("rm %s", data.Destination.ValueString())); err != nil {
	//	resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read file, got error: %s", err))
	//} else {
	//	data.Result = types.StringValue("")
	//}

	data.Result = types.StringValue("")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
