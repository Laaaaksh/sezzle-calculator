package boot_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/boot"
)

type InitializeTestSuite struct {
	suite.Suite
	ctx context.Context
	dir string
}

func TestInitializeTestSuite(t *testing.T) {
	suite.Run(t, new(InitializeTestSuite))
}

func (s *InitializeTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.dir = s.T().TempDir()
	s.writeFile(fileDefault, fixtureDefaultToml)
	s.writeFile(fileTest, fixtureTestToml)
	s.T().Setenv(envVarEnv, envNameTest)
	s.T().Setenv(envVarConfigDir, s.dir)
	s.T().Setenv(envVarPort, envValueEmpty)
}

func (s *InitializeTestSuite) TearDownTest() {
}

func (s *InitializeTestSuite) writeFile(name, content string) {
	s.Require().NoError(os.WriteFile(filepath.Join(s.dir, name), []byte(content), 0o600))
}

func (s *InitializeTestSuite) get(app *boot.App, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func (s *InitializeTestSuite) TestInitializeBuildsAppThatServesTheOperationsList() {
	app, err := boot.Initialize(s.ctx)

	s.Require().NoError(err)
	rec := s.get(app, testOperationsPath)
	s.Equal(http.StatusOK, rec.Code)
	s.Contains(rec.Body.String(), wantOperationsBodyStart)
}

func (s *InitializeTestSuite) TestInitializeBuildsAppThatServesTheLiveProbe() {
	app, err := boot.Initialize(s.ctx)

	s.Require().NoError(err)
	s.Equal(http.StatusOK, s.get(app, testLivePath).Code)
}

func (s *InitializeTestSuite) TestInitializeBuildsAppThatServesLive() {
	app, err := boot.Initialize(s.ctx)

	s.Require().NoError(err)
	rec := s.get(app, testLivePath)
	s.Equal(http.StatusOK, rec.Code)
	s.JSONEq(wantLiveBody, rec.Body.String())
}

func (s *InitializeTestSuite) TestInitializeMarksAppReadyAfterRouterSetup() {
	app, err := boot.Initialize(s.ctx)

	s.Require().NoError(err)
	rec := s.get(app, testReadyPath)
	s.Equal(http.StatusOK, rec.Code)
	s.JSONEq(wantReadyBody, rec.Body.String())
}

func (s *InitializeTestSuite) TestInitializeWithMissingConfigDirFailsAtConfigStage() {
	s.T().Setenv(envVarConfigDir, filepath.Join(s.dir, missingDirName))

	app, err := boot.Initialize(s.ctx)

	s.Nil(app)
	s.ErrorIs(err, boot.ErrConfigLoad)
	s.NotErrorIs(err, boot.ErrLoggerInit)
}

func (s *InitializeTestSuite) TestInitializeWithMalformedConfigFailsAtConfigStage() {
	s.writeFile(fileDefault, fixtureBadToml)

	app, err := boot.Initialize(s.ctx)

	s.Nil(app)
	s.ErrorIs(err, boot.ErrConfigLoad)
}

func (s *InitializeTestSuite) TestInitializeErrorTextIsTheBareCause() {
	s.T().Setenv(envVarConfigDir, filepath.Join(s.dir, missingDirName))

	_, err := boot.Initialize(s.ctx)

	s.NotContains(err.Error(), boot.ErrConfigLoad.Error())
}
