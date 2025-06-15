package main

import (
	"errors"
	"net/http"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func vkHandlerEcho(c echo.Context, server *Server) error {
	groupName := c.Param("group")

	data, err := server.Vk(c.Request().Context(), groupName)

	if err != nil {
		switch {
		case errors.Is(err, ErrorGroupNotAllowed):
			return echo.NewHTTPError(http.StatusForbidden)
		case errors.Is(err, ErrorGroupNotFound):
			return echo.NewHTTPError(http.StatusNotFound)
		case errors.Is(err, ErrorInvalidGroupName):
			return echo.NewHTTPError(http.StatusBadRequest)
		default:
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	}

	return c.Blob(http.StatusOK, "application/xml; charset=utf-8", data)
}

func settingsSaltHandlerEcho(c echo.Context, server *Server) error {
	return c.String(http.StatusOK, server.SettingsSalt())
}

func settingsUpdateHandlerEcho(c echo.Context, server *Server) error {
	var input struct {
		Data string `json:"data"`
		Hash string `json:"hash"`
		Salt string `json:"salt"`
	}

	err := c.Bind(&input)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	data, err := server.SettingsUpdate(input.Data, input.Hash, input.Salt)

	if err != nil {
		switch {
		case errors.Is(err, ErrorInvalidPassword):
			return echo.NewHTTPError(http.StatusUnauthorized)
		default:
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	}

	return c.String(http.StatusOK, data)
}

func errorHandlerEcho(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}
	var httpError *echo.HTTPError
	if errors.As(err, &httpError) {
		c.NoContent(httpError.Code)
	} else {
		c.NoContent(http.StatusInternalServerError)
	}
}

func startServerEcho(server *Server) {
	e := echo.New()
	e.DisableHTTP2 = true
	e.HideBanner = true
	e.HidePort = true
	e.HTTPErrorHandler = errorHandlerEcho

	metricsSkipper := func(c echo.Context) bool { return c.Path() == "/metrics" }
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{Skipper: metricsSkipper}))
	e.Use(middleware.Recover())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(3))) // 3 RPS per ctx.RealIP

	e.Use(echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{
		Subsystem: "myapp",
		Skipper:   metricsSkipper,
	}))

	e.GET("/metrics", echoprometheus.NewHandler())

	e.GET("/vk/:group", func(c echo.Context) error {
		return vkHandlerEcho(c, server)
	})

	e.GET("/myapp_settings/salt", func(c echo.Context) error {
		return settingsSaltHandlerEcho(c, server)
	})
	e.POST("/myapp_settings/update", func(c echo.Context) error {
		return settingsUpdateHandlerEcho(c, server)
	})
	e.File("/myapp_settings", "/settings.html")

	e.Logger.Fatal(e.Start(":8080"))
}
