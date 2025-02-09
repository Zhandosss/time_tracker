package handlers

import (
	"context"
	"effectiveMobileTestProblem/internal/model"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
)

type PaginationInfo struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

//go:generate go run github.com/vektra/mockery/v2@v2.43.2 --name UserService --output mocks/
type UserService interface {
	CreateUser(ctx context.Context, user *model.User) (string, error)
	GetUserById(ctx context.Context, id string) (*model.User, error)
	GetUserByPassport(ctx context.Context, passport string) (*model.User, error)
	GetUsers(ctx context.Context, filterAndPagination *model.FilterAndPagination) ([]*model.User, error)
	DeleteUserById(ctx context.Context, id string) error
	DeleteUserByPassport(ctx context.Context, passport string) error
	UpdateUserById(ctx context.Context, id string, user *model.User) error
	UpdateUserByPassport(ctx context.Context, passport string, user *model.User) error
}

type WorkService interface {
	StartWork(ctx context.Context, work *model.Work) (string, error)
	StopWork(ctx context.Context, id string) error
	GetWorkById(ctx context.Context, id string) (*model.Work, error)
	GetWorks(ctx context.Context, user string) ([]*model.Work, error)
}

type Handlers struct {
	UserService UserService
	WorkService WorkService
}

//@title Time Tracker API
//@version 1.0
//@description This is a simple API for tracking time spent on work for users

// @basePath /api/
func NewHandlers(e *echo.Echo, userSer UserService, workSer WorkService) {
	handlers := &Handlers{
		UserService: userSer,
		WorkService: workSer,
	}

	e.Use(middleware.RequestID())

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:       true,
		LogStatus:    true,
		LogRequestID: true,
		LogRemoteIP:  true,
		LogError:     true,
		LogLatency:   true,
		LogMethod:    true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log.Info().
				Str("latentcy", v.Latency.String()).
				Str("requestID", v.RequestID).
				Err(v.Error).
				Str("remoteIP", v.RemoteIP).
				Str("URI", v.URI).
				Int("status", v.Status).
				Str("method", v.Method).
				Msg("request")

			return nil
		},
	}))
	e.Use(middleware.Recover())

	api := e.Group("/api")
	{
		user := api.Group("/user")
		{
			user.POST("", handlers.CreateUser)
			user.GET("", handlers.GetUsers)
			user.GET("/id/:id", handlers.GetUserById)
			user.GET("/passport/:passport", handlers.GetUserByPassport)
			user.DELETE("/id/:id", handlers.DeleteUserById)
			user.DELETE("/passport/:passport", handlers.DeleteUserByPassport)
			user.PATCH("/id/:id", handlers.UpdateUserById)
			user.PATCH("/passport/:passport", handlers.UpdateUserByPassport)
		}
		work := api.Group("/work")
		{
			work.POST("", handlers.StartWork)
			work.GET("/:id", handlers.GetWorkById)
			work.GET("", handlers.GetWorks)
			work.GET("/:id/stop", handlers.StopWork)
		}
	}
}
