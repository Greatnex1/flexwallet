package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
)

func main() {
	ex := NewExchange()
	e := echo.New()
	e.HTTPErrorHandler = httpErrorHandler
	e.GET("/book/:market", ex.handleGetBook)
	e.POST("/order", ex.handlePlaceOrder)
	e.DELETE("/order/:id", ex.handleCancelOrder)
	err := e.Start(":3000")
	if err != nil {
		return
	}
	fmt.Println("Go running ....")
}
