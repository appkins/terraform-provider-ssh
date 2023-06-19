package log

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func Info(ctx context.Context, message string, data ...any) {
	tflog.Info(ctx, fmt.Sprintf(message, data...))
}

func Debug(ctx context.Context, message string, data ...any) {
	tflog.Debug(ctx, fmt.Sprintf(message, data...))
}
