package api

import (
	"github.com/labstack/echo"
	"net/http"
)

type (
	//ErrResponse main struct for error handling
	ErrResponse struct {
		Error ErrContent `json:"error"`
	}

	//ErrContent contains the code and message of the Error
	ErrContent struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	//ErrResponseValidation contains the type of error found and its fields and correspondent messages
	ErrResponseValidation struct {
		Type   string           `json:"error_type"`
		Errors []*ErrValidation `json:"errors"`
	}

	//ErrValidation contains the field and correspondent error message
	ErrValidation struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	}
)

//Error handler for api errors
func Error(err error, c echo.Context) {
	code := http.StatusServiceUnavailable
	msg := http.StatusText(code)

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = he.Message.(string)
	}

	if c.Echo().Debug {
		msg = err.Error()
	}

	content := map[string]interface{}{
		"id":      c.Response().Header().Get(echo.HeaderXRequestID),
		"message": msg,
		"status":  code,
	}

	c.Logger().Errorj(content)

	if !c.Response().Committed {
		if c.Request().Method == echo.HEAD {
			c.NoContent(code)
		} else {
			c.JSON(code, &ErrResponse{ErrContent{code, msg}})
		}
	}
}
