package integrationtests

import (
	"bytes"
	"context"
	"database/sql"
	"github.com/Spear5030/YAGopherMart/internal/app"
	"github.com/Spear5030/YAGopherMart/internal/config"
	"github.com/Spear5030/YAGopherMart/testutils"
	"github.com/Spear5030/YAGopherMart/testutils/fixtureloader"
	"github.com/Spear5030/YAGopherMart/testutils/testcontainer"
	"github.com/go-testfixtures/testfixtures/v3"
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

func (s *TestSuite) TestGetOrders() {
	//jwt for test_name user with id 10001
	auth := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJFeHBpcmVkQXQiOiIyMDIzLTAxLTI1VDAwOjUwOjMwLjE5MDg4NjIrMDM6MDAiLCJpZCI6MTAwMDF9.m6F9WuPxEuMaugGZN9gTixmB0iiUNhMUjPKPqQkzExQ"
	r, err := http.NewRequest(http.MethodGet, s.server.URL+"/api/user/orders", nil)
	s.Require().NoError(err)
	r.Header.Add("Authorization", auth)
	resp, err := s.server.Client().Do(r)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)
	expected := s.fixtures.LoadString(s.T(), "fixtures/api/getOrders.json")
	testutils.JSONEq(s.T(), expected, resp.Body)
}

// это как я понимаю, уже больше end 2 end тест
func (s *TestSuite) TestPostOrder() {
	b := bytes.NewBufferString("{\"login\": \"login\",\"password\": \"password\"}")

	r, err := http.NewRequest(http.MethodPost, s.server.URL+"/api/user/register", b)
	s.Require().NoError(err)
	resp, err := s.server.Client().Do(r)
	s.Require().NoError(err)
	defer resp.Body.Close()
	auth := resp.Header.Get("Authorization")

	b = bytes.NewBufferString("79927398713")
	r, err = http.NewRequest(http.MethodPost, s.server.URL+"/api/user/orders", b)
	s.Require().NoError(err)

	resp, err = s.server.Client().Do(r)
	s.Require().NoError(err)

	defer resp.Body.Close()
	s.Require().Equal(http.StatusUnauthorized, resp.StatusCode)

	b = bytes.NewBufferString("79927398713")
	r, err = http.NewRequest(http.MethodPost, s.server.URL+"/api/user/orders", b)
	s.Require().NoError(err)
	r.Header.Add("Authorization", auth)

	resp, err = s.server.Client().Do(r)
	s.Require().NoError(err)

	defer resp.Body.Close()
	s.Require().Equal(http.StatusAccepted, resp.StatusCode)

	b = bytes.NewBufferString("79927398713")
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

func (s *TestSuite) SetupTest() {
	db, err := sql.Open("postgres", s.postgresContainer.GetDSN())
	s.Require().NoError(err)

	defer db.Close()

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.FS(testutils.Fixtures),
		testfixtures.Directory("fixtures/storage"),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())
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

	s.fixtures = fixtureloader.NewLoader(testutils.Fixtures)

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
