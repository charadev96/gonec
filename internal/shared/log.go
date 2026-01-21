package shared

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

func Logger() zerolog.Logger {
	output := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "15:04",
	}
	output.FormatFieldName = func(i any) string {
		return fmt.Sprintf("\n         \033[30m- \033[36m%s: \033[0m", i)
	}
	output.FormatErrFieldName = func(i any) string {
		return fmt.Sprintf("\n         \033[30m- \033[31m%s: \033[0m", i)
	}

	logger := zerolog.New(output).
		With().
		Timestamp().
		Logger()
	return logger
}
