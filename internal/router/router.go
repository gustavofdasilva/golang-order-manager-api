package router

import "github.com/labstack/echo/v4"

func InitRouter(e *echo.Group) {
	initUserRouter(e)
	initAuthRouter(e)
}
