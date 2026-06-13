package log

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func Info(ctx context.Context, format string, data ...any) {
	tflog.Info(ctx, fmt.Sprintf(format, data...))
}

func Debug(ctx context.Context, format string, data ...any) {
	tflog.Debug(ctx, fmt.Sprintf(format, data...))
}

func Error(ctx context.Context, format string, data ...any) {
	tflog.Error(ctx, fmt.Sprintf(format, data...))
}
