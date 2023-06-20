package remote

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/appkins/terraform-provider-ssh/internal/log"
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

func (p *Provisioner) CopyFiles(ctx context.Context, createFiles []File, config *Config) error {

	ssh, _ := p.CreateClient(config)

	retryDelay := p.RetryDelay

	if config.RetryDelay != 0 {
		retryDelay = config.RetryDelay
	}

	for _, f := range createFiles {
		copyFile := func(f File) error {
			if !f.Source.IsUnknown() {
				src, srcErr := os.Open(f.Source.String())
				if srcErr != nil {
					log.Debug(ctx, "Failed to open source file %s: %v\n", f.Source, srcErr)
					return srcErr
				}
				srcStat, statErr := src.Stat()
				if statErr != nil {
					log.Debug(ctx, "Failed to stat source file %s: %v\n", f.Source, statErr)
					_ = src.Close()
					return statErr
				}
				_ = ssh.WriteFile(src, srcStat.Size(), f.Destination.String())
				log.Debug(ctx, "Copied %s to remote file %s:%s: %d bytes\n", f.Source, ssh.Server, f.Destination, srcStat.Size())
				_ = src.Close()
			} else {
				buffer := bytes.NewBufferString(f.Content.String())
				if err := ssh.WriteFile(buffer, int64(buffer.Len()), f.Destination.String()); err != nil {
					log.Debug(ctx, "Failed to copy content to remote file %s:%s:%s: %v\n", ssh.Server, ssh.Port, f.Destination, err)
					return err
				}
				log.Debug(ctx, "Created remote file %s:%s:%s: %d bytes\n", ssh.Server, ssh.Port, f.Destination, len(f.Content.String()))
			}
			// Permissions change
			if !f.Permissions.IsUnknown() {
				outStr, errStr, _, err := ssh.Run(fmt.Sprintf("chmod %s \"%s\"", f.Permissions, f.Destination))
				log.Debug(ctx, "Permissions file %s:%s: %v %v\n", f.Destination, f.Permissions, outStr, errStr)
				if err != nil {
					return err
				}
			}
			// Owner
			if !f.Owner.IsUnknown() {
				outStr, errStr, _, err := ssh.Run(fmt.Sprintf("chown %s \"%s\"", f.Owner, f.Destination))
				log.Debug(ctx, "Owner file %s:%s: %v %v\n", f.Destination, f.Owner, outStr, errStr)
				if err != nil {
					return err
				}
			}
			// Group
			if !f.Group.IsUnknown() {
				outStr, errStr, _, err := ssh.Run(fmt.Sprintf("chgrp %s \"%s\"", f.Group, f.Destination))
				log.Debug(ctx, "Group file %s:%s: %v %v\n", f.Destination, f.Group, outStr, errStr)
				if err != nil {
					return err
				}
			}
			return nil
		}
		for {
			err := copyFile(f)
			if err == nil {
				break
			}
			select {
			case <-time.After(retryDelay):
			// Retry
			case <-ctx.Done():
				return fmt.Errorf("%s: %w", ctx.Err(), err)
			}
		}
	}
	return nil
}
