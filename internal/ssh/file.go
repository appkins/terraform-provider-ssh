package ssh

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type File struct {
	Source      types.String `tfsdk:"source"`
	Destination types.String `tfsdk:"destination"`
	Content     types.String `tfsdk:"content"`
	Permissions types.String `tfsdk:"permissions"`
	Owner       types.String `tfsdk:"owner"`
	Group       types.String `tfsdk:"group"`
}

func (c *Client) WriteFileToHost(file File) error {
	dest := file.Destination.ValueString()
	if dest == "" {
		return fmt.Errorf("destination is required")
	}
	_, _, _, err := c.Run("mkdir -p "+dest, time.Duration(30*time.Second))
	if err != nil {
		return err
	}

	if src := file.Source.ValueString(); src != "" {
		return c.Scp(src, file.Destination.ValueString())
	}

	if content := file.Content.ValueString(); content != "" {
		reader := strings.NewReader(content)
		size := int64(len(content))
		return c.WriteFile(reader, size, file.Destination.ValueString())
	}

	return fmt.Errorf("source or content is required")
}
