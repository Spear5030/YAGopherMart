package handler

import (
	"context"
	"database/sql"
	"github.com/Spear5030/YAGopherMart/db/migrate"
	"github.com/Spear5030/YAGopherMart/internal/app"
	"github.com/Spear5030/YAGopherMart/internal/config"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/suite"
	"github.com/underbek/examples-go/integration_tests/testentities"
	"github.com/underbek/examples-go/testutils"
	"net/http/httptest"
	"time"
)

type TestSuite struct {
	suite.Suite
	app       *app.App
	server    *httptest.Server
	container *testutils.PostgreSQLContainer
	fixtures  *testutils.FixtureLoader
}

func (s *TestSuite) SetupSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer ctxCancel()

	c, err := testutils.NewPostgreSQLContainer(ctx)
	s.Require().NoError(err)

	s.container = c

	s.fixtures = testutils.NewFixtureLoader(s.T(), testentities.Fixtures)

	db, err := sql.Open("postgres", c.GetDSN())
	s.Require().NoError(err)

	err = testutils.Migrate(db, migrate.Migrations)
	s.Require().NoError(err)

	cfg, err := config.New()
	s.Require().NoError(err)
	s.app, err = app.New(cfg)
	s.Require().NoError(err)

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("../testentities/fixtures/postgres"),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())

	s.server = httptest.NewServer(s.app.HTTPServer.Handler)
}

func (s *TestSuite) TearDownSuite() {
	_, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()
}
