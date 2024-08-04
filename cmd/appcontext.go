package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// AppContext echo context extended with application specific fields
type AppContext struct {
	echo.Context
	Config Config
}

type ErrorRes struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

// OkRes to deprecate. No reason in sending this struct, there is already HTTP Code 2xx for that
type OkRes struct {
	Ok bool `json:"ok"`
}

func (c *AppContext) validateAndBindRequest(r interface{}) error {

	if err := c.Bind(r); err != nil {
		return err
	}

	if err := c.Validate(r); err != nil {
		return c.errorResponse(err.Error(), "910")
	}

	return nil
}

// to deprecate. No reason in sending OkRes struct, there is already HTTP Code 2xx for that
func (c *AppContext) okResponse() error {
	return c.JSON(http.StatusOK, OkRes{Ok: true})
}

func (c *AppContext) okResponseWithData(response interface{}) error {
	return c.JSON(http.StatusOK, response)

}

func (c *AppContext) errorResponse(error string, code string) error {
	return c.JSON(http.StatusBadRequest, ErrorRes{Detail: error, Code: code})

}
