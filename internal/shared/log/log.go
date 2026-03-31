package log

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/rs/zerolog"
)

func NewLogger(module string) zerolog.Logger {
	out := zerolog.ConsoleWriter{
		Out:           os.Stderr,
		TimeFormat:    "15:04",
		PartsOrder:    []string{"time", "level", "module", "message"},
		FieldsExclude: []string{"module"},
	}

	out.FormatPartValueByName = func(i any, s string) string {
		if s == "module" && i != nil {
			return strings.ToUpper(fmt.Sprintf("%s", i))
		}
		return ""
	}

	out.FormatFieldName = func(i any) string {
		return fmt.Sprintf("\n         \033[30m- \033[36m%s: \033[0m", i)
	}

	out.FormatErrFieldName = func(i any) string {
		return fmt.Sprintf("\n         \033[30m- \033[31m%s: \033[0m", i)
	}

	logger := zerolog.New(out).
		With().
		Timestamp().
		Str("module", module).
		Logger()

	return logger
}

func NewInterceptor(l zerolog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l := l.With().Fields(fields).Logger()
		switch lvl {
		case logging.LevelDebug:
			l.Debug().Msg(msg)
		case logging.LevelInfo:
			l.Info().Msg(msg)
		case logging.LevelWarn:
			l.Warn().Msg(msg)
		case logging.LevelError:
			l.Error().Msg(msg)
		}
	})
}
