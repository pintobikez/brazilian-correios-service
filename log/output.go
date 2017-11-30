package log

import (
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"io"
	"os"
)

//File Retrieve the io.Writer from a file if exists, otherwise returns a os.Stdout
func File(filePath string) io.Writer {
	file, err := os.OpenFile(
		filePath,
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0666,
	)

	if err != nil {
		return os.Stdout
	}

	return file
}

//LoggerWithOutput Creates a logger without output Middleware to echo
func LoggerWithOutput(w io.Writer) echo.MiddlewareFunc {
	config := mw.DefaultLoggerConfig
	config.Output = w

	return mw.LoggerWithConfig(config)
}
