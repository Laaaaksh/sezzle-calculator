package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/config"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
)

type LoadTestSuite struct {
	suite.Suite
	dir string
}

func TestLoadTestSuite(t *testing.T) {
	suite.Run(t, new(LoadTestSuite))
}

func (s *LoadTestSuite) SetupTest() {
	s.dir = s.T().TempDir()
	s.writeFile(fileDefault, fixtureDefaultToml)
	s.writeFile(fileDev, fixtureDevToml)
	s.writeFile(fileTest, fixtureTestToml)
	s.writeFile(fileProd, fixtureProdToml)

	// Blank every variable Load reads so a value exported in the developer's
	// shell cannot leak into a test.
	for _, name := range []string{
		envVarEnvironment, envVarPort, envVarReadTimeout, envVarMaxBody,
		envVarCORSOrigins, envVarDecimalPrec,
	} {
		s.T().Setenv(name, envValueEmpty)
	}
	s.T().Setenv(envVarConfigDir, s.dir)
}

func (s *LoadTestSuite) writeFile(name, content string) {
	s.Require().NoError(os.WriteFile(filepath.Join(s.dir, name), []byte(content), 0o600))
}

func (s *LoadTestSuite) TestLoadWhenEnvUnsetUsesDevEnvironment() {
	cfg, err := config.Load()

	s.Require().NoError(err)
	s.Equal(constants.EnvDev, cfg.App.Env)
	s.Equal(wantDevPort, cfg.Server.Port)
}

func (s *LoadTestSuite) TestLoadKeepsDefaultsNotOverriddenByEnvFile() {
	cfg, err := config.Load()

	s.Require().NoError(err)
	s.Equal(wantDefaultReadTimeout, cfg.Server.ReadTimeout)
	s.Equal(wantDefaultWriteTimeout, cfg.Server.WriteTimeout)
	s.Equal(wantDefaultIdleTimeout, cfg.Server.IdleTimeout)
	s.Equal(wantDefaultGracePeriod, cfg.Server.ShutdownGracePeriod)
	s.Equal(wantDefaultMaxBody, cfg.Server.MaxBodyBytes)
	s.Equal([]string{wantDefaultOrigin}, cfg.CORS.AllowedOrigins)
	s.Equal(wantDefaultMaxOperand, cfg.Calculator.MaxOperandLength)
	s.Equal(wantDefaultMaxExponent, cfg.Calculator.MaxExponent)
	s.Equal(wantDefaultDecimalPrec, cfg.Calculator.DecimalPrecision)
}

func (s *LoadTestSuite) TestLoadWhenEnvIsDevAppliesDevOverrides() {
	s.T().Setenv(envVarEnvironment, envNameDev)

	cfg, err := config.Load()

	s.Require().NoError(err)
	s.Equal(constants.EnvDev, cfg.App.Env)
	s.Equal(wantDevPort, cfg.Server.Port)
}

func (s *LoadTestSuite) TestLoadWhenEnvIsTestAppliesTestOverrides() {
	s.T().Setenv(envVarEnvironment, envNameTest)

	cfg, err := config.Load()

	s.Require().NoError(err)
	s.Equal(constants.EnvTest, cfg.App.Env)
	s.Equal(wantTestPort, cfg.Server.Port)
	s.Equal(wantTestReadTimeout, cfg.Server.ReadTimeout)
	s.Equal([]string{wantTestOriginOne, wantTestOriginTwo}, cfg.CORS.AllowedOrigins)
}

func (s *LoadTestSuite) TestLoadWhenEnvIsProdAppliesProdOverrides() {
	s.T().Setenv(envVarEnvironment, envNameProd)

	cfg, err := config.Load()

	s.Require().NoError(err)
	s.Equal(constants.EnvProd, cfg.App.Env)
	s.Equal(wantProdPort, cfg.Server.Port)
	s.Equal(wantProdMaxExponent, cfg.Calculator.MaxExponent)
	s.Equal(wantDefaultMaxOperand, cfg.Calculator.MaxOperandLength)
}

func (s *LoadTestSuite) TestLoadWhenScalarEnvVarSetOverridesFiles() {
	s.T().Setenv(envVarPort, envValuePort)
	s.T().Setenv(envVarMaxBody, envValueMaxBody)
	s.T().Setenv(envVarDecimalPrec, envValueDecimalPrec)

	cfg, err := config.Load()

	s.Require().NoError(err)
	s.Equal(wantOverridePort, cfg.Server.Port)
	s.Equal(wantOverrideMaxBody, cfg.Server.MaxBodyBytes)
	s.Equal(wantOverrideDecimalPrec, cfg.Calculator.DecimalPrecision)
}

func (s *LoadTestSuite) TestLoadWhenDurationEnvVarSetOverridesFiles() {
	s.T().Setenv(envVarReadTimeout, envValueReadTimeout)

	cfg, err := config.Load()

	s.Require().NoError(err)
	s.Equal(wantOverrideReadTimeout, cfg.Server.ReadTimeout)
}

func (s *LoadTestSuite) TestLoadWhenCORSEnvVarSetSplitsCommaSeparatedList() {
	s.T().Setenv(envVarCORSOrigins, envValueCORSOrigins)

	cfg, err := config.Load()

	s.Require().NoError(err)
	s.Equal([]string{wantOverrideOriginOne, wantOverrideOriginTwo}, cfg.CORS.AllowedOrigins)
}

func (s *LoadTestSuite) TestLoadWhenEnvVarSetBeatsEnvFile() {
	s.T().Setenv(envVarEnvironment, envNameTest)
	s.T().Setenv(envVarPort, envValuePort)

	cfg, err := config.Load()

	s.Require().NoError(err)
	s.Equal(wantOverridePort, cfg.Server.Port)
}

func (s *LoadTestSuite) TestLoadWhenEnvIsUnknownReturnsError() {
	s.T().Setenv(envVarEnvironment, envNameUnknown)

	cfg, err := config.Load()

	s.Nil(cfg)
	s.ErrorIs(err, config.ErrUnknownEnvironment)
	s.Contains(err.Error(), wantUnknownEnvMessage)
}

func (s *LoadTestSuite) TestLoadWhenDefaultFileMissingReturnsError() {
	s.Require().NoError(os.Remove(filepath.Join(s.dir, fileDefault)))

	cfg, err := config.Load()

	s.Nil(cfg)
	s.Require().Error(err)
	s.Contains(err.Error(), fileDefault)
}

func (s *LoadTestSuite) TestLoadWhenEnvFileMissingReturnsError() {
	s.Require().NoError(os.Remove(filepath.Join(s.dir, fileDev)))

	cfg, err := config.Load()

	s.Nil(cfg)
	s.Require().Error(err)
	s.Contains(err.Error(), fileDev)
}

func (s *LoadTestSuite) TestLoadWhenConfigDirUnsetFallsBackToDefaultDir() {
	s.T().Setenv(envVarConfigDir, envValueEmpty)

	cfg, err := config.Load()

	// The test runs from the package directory, which has no config/ folder.
	s.Nil(cfg)
	s.Require().Error(err)
	s.Contains(err.Error(), filepath.Join(envValueDefaultNoDir, fileDefault))
}

func (s *LoadTestSuite) TestLoadWhenConfigDirDoesNotExistReturnsError() {
	s.T().Setenv(envVarConfigDir, envValueMissingDir)

	cfg, err := config.Load()

	s.Nil(cfg)
	s.Require().Error(err)
	s.Contains(err.Error(), envValueMissingDir)
}

func (s *LoadTestSuite) TestLoadWhenDefaultFileMalformedReturnsError() {
	s.writeFile(fileDefault, fixtureMalformedToml)

	cfg, err := config.Load()

	s.Nil(cfg)
	s.Require().Error(err)
	s.Contains(err.Error(), fileDefault)
}

func (s *LoadTestSuite) TestLoadWhenEnvFileMalformedReturnsError() {
	s.writeFile(fileDev, fixtureMalformedToml)

	cfg, err := config.Load()

	s.Nil(cfg)
	s.Require().Error(err)
	s.Contains(err.Error(), fileDev)
}

func (s *LoadTestSuite) TestLoadWhenFileHasUnknownKeyReturnsError() {
	s.writeFile(fileDev, fixtureUnknownKeyToml)

	cfg, err := config.Load()

	s.Nil(cfg)
	s.Require().Error(err)
	s.Contains(err.Error(), "prot")
}

func (s *LoadTestSuite) TestLoadWhenDurationIsUnparsableReturnsError() {
	s.writeFile(fileDev, fixtureBadDurationToml)

	cfg, err := config.Load()

	s.Nil(cfg)
	s.Require().Error(err)
	s.Contains(err.Error(), "read_timeout")
}

func (s *LoadTestSuite) TestLoadWhenFileValueFailsValidationReturnsNamedKey() {
	s.writeFile(fileDev, fixtureInvalidValueToml)

	cfg, err := config.Load()

	s.Nil(cfg)
	s.ErrorIs(err, config.ErrOperandLengthOutOfRange)
	s.Contains(err.Error(), wantOperandMessage)
}

func (s *LoadTestSuite) TestLoadWhenEnvVarFailsValidationReturnsNamedKey() {
	s.T().Setenv(envVarPort, envValueZero)

	cfg, err := config.Load()

	s.Nil(cfg)
	s.ErrorIs(err, config.ErrPortOutOfRange)
	s.Contains(err.Error(), wantPortMessage)
}
