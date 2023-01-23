package integrationtests

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Spear5030/YAGopherMart/internal/app"
	"github.com/Spear5030/YAGopherMart/internal/config"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	//"github.com/underbek/integration-test-the-best/testutils/testcontainer"
	// при подключении ловил панику sql: Register called twice for driver pgx
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
	postgresContainer *TestDatabase
	//fixtures  *testutils.FixtureLoader
}

type TestDatabase struct {
	instance testcontainers.Container
}

func NewTestDatabase() *TestDatabase {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:12",
		ExposedPorts: []string{"5432/tcp"},
		AutoRemove:   true,
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "postgres",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	postgres, _ := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	return &TestDatabase{
		instance: postgres,
	}
}

func (db *TestDatabase) Port() int {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	p, _ := db.instance.MappedPort(ctx, "5432")
	return p.Int()
}

func (db *TestDatabase) ConnectionString() string {
	return fmt.Sprintf("postgres://postgres:postgres@127.0.0.1:%d/postgres", db.Port())
}

func (db *TestDatabase) Close(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	require.NoError(t, db.instance.Terminate(ctx))
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
	_, ctxCancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer ctxCancel()

	var err error

	s.postgresContainer = NewTestDatabase()

	s.app, err = app.New(
		config.Config{
			Database: s.postgresContainer.ConnectionString(),
			Key:      "P4SSW0RD",
		})
	s.Require().NoError(err)
	s.server = httptest.NewServer(s.app.HTTPServer.Handler)
}

func TestSuite_PostgreSQLStorage(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TearDownSuite() {
	_, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()
}
