package integrationtests

import (
	"bytes"
	"context"
	"github.com/Spear5030/YAGopherMart/internal/app"
	"github.com/Spear5030/YAGopherMart/internal/config"
	"github.com/Spear5030/YAGopherMart/testutils/fixtureloader"
	"github.com/Spear5030/YAGopherMart/testutils/testcontainer"

	"net/http"

	"testing"

	"github.com/stretchr/testify/suite"
	"net/http/httptest"
	"time"
)

type TestSuite struct {
	suite.Suite
	app               *app.App
	server            *httptest.Server
	postgresContainer *testcontainer.PostgresContainer
	fixtures          *fixtureloader.Loader
}

func (s *TestSuite) TestRegisterUser() {
	b := bytes.NewBufferString("{\"login\": \"login\",\"password\": \"password\"}")

	r, err := http.NewRequest(http.MethodPost, s.server.URL+"/api/user/register", b)
	s.Require().NoError(err)
	resp, err := s.server.Client().Do(r)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	auth := resp.Header.Get("Authorization")
	s.Require().Contains(auth, "Bearer ", "No Authorization")

	b = bytes.NewBufferString("{\"login\": \"login\",\"password\": \"password\"}")
	r, err = http.NewRequest(http.MethodPost, s.server.URL+"/api/user/register", b)
	s.Require().NoError(err)
	resp, err = s.server.Client().Do(r)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusConflict, resp.StatusCode)
}

func (s *TestSuite) TestPostOrder() {
	b := bytes.NewBufferString("{\"login\": \"login2\",\"password\": \"password\"}")

	r, err := http.NewRequest(http.MethodPost, s.server.URL+"/api/user/register", b)
	s.Require().NoError(err)
	resp, err := s.server.Client().Do(r)
	s.Require().NoError(err)
	defer resp.Body.Close()
	auth := resp.Header.Get("Authorization")

	b = bytes.NewBufferString("12345678903")
	r, err = http.NewRequest(http.MethodPost, s.server.URL+"/api/user/orders", b)
	s.Require().NoError(err)

	resp, err = s.server.Client().Do(r)
	s.Require().NoError(err)

	defer resp.Body.Close()
	s.Require().Equal(http.StatusUnauthorized, resp.StatusCode)

	b = bytes.NewBufferString("12345678903")
	r, err = http.NewRequest(http.MethodPost, s.server.URL+"/api/user/orders", b)
	s.Require().NoError(err)
	r.Header.Add("Authorization", auth)

	resp, err = s.server.Client().Do(r)
	s.Require().NoError(err)

	defer resp.Body.Close()
	s.Require().Equal(http.StatusAccepted, resp.StatusCode)

	b = bytes.NewBufferString("12345678903")
	r, err = http.NewRequest(http.MethodPost, s.server.URL+"/api/user/orders", b)
	s.Require().NoError(err)
	r.Header.Add("Authorization", auth)

	resp, err = s.server.Client().Do(r)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	b = bytes.NewBufferString("12345678904")
	r, err = http.NewRequest(http.MethodPost, s.server.URL+"/api/user/orders", b)
	s.Require().NoError(err)
	r.Header.Add("Authorization", auth)

	resp, err = s.server.Client().Do(r)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusUnprocessableEntity, resp.StatusCode)
}

func (s *TestSuite) SetupSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer ctxCancel()

	var err error

	s.postgresContainer, err = testcontainer.NewPostgresContainer(ctx)

	s.app, err = app.New(
		config.Config{
			Database: s.postgresContainer.GetDSN(),
			Key:      "P4SSW0RD",
		})
	s.Require().NoError(err)
	s.server = httptest.NewServer(s.app.HTTPServer.Handler)
}

func TestSuite_PostgreSQLStorage(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TearDownSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	s.Require().NoError(s.postgresContainer.Terminate(ctx))
}
