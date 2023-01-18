package handler

import (
	"context"
	"github.com/Spear5030/YAGopherMart/internal/app"
	"github.com/Spear5030/YAGopherMart/internal/config"
	"github.com/stretchr/testify/suite"
	"net/http/httptest"
	"time"
)

type TestSuite struct {
	suite.Suite
	app    *app.App
	server *httptest.Server
}

func (s *TestSuite) SetupSuite() {
	_, ctxCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer ctxCancel()

	cfg, err := config.New()
	s.Require().NoError(err)
	s.app, err = app.New(cfg)
	s.Require().NoError(err)

	s.server = httptest.NewServer(s.app.HTTPServer.Handler)
}

func (s *TestSuite) TearDownSuite() {
	_, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()
}
