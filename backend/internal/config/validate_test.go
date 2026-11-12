package config_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/config"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
)

type ValidateTestSuite struct {
	suite.Suite
	cfg config.Config
}

func TestValidateTestSuite(t *testing.T) {
	suite.Run(t, new(ValidateTestSuite))
}

// SetupTest rebuilds a known-good config so each test breaks exactly one rule.
func (s *ValidateTestSuite) SetupTest() {
	s.cfg = config.Config{
		App: config.AppConfig{Env: constants.EnvDev},
		Server: config.ServerConfig{
			Port:                validPort,
			ReadTimeout:         validTimeout,
			WriteTimeout:        validTimeout,
			IdleTimeout:         validTimeout,
			RequestTimeout:      validRequestTimeout,
			ShutdownGracePeriod: validTimeout,
			MaxBodyBytes:        validMaxBody,
		},
		CORS: config.CORSConfig{AllowedOrigins: []string{validOrigin}},
		Calculator: config.CalculatorConfig{
			MaxOperandLength: validMaxOperand,
			MaxExponent:      validMaxExponent,
			DecimalPrecision: validDecimalPrecision,
		},
	}
}

func (s *ValidateTestSuite) TestValidateWhenConfigIsValidReturnsNil() {
	s.NoError(s.cfg.Validate())
}

func (s *ValidateTestSuite) TestValidateAcceptsOriginWithPortAndTrailingSlash() {
	s.cfg.CORS.AllowedOrigins = []string{validOriginWithPort, validOriginRootSlash}

	s.NoError(s.cfg.Validate())
}

func (s *ValidateTestSuite) TestValidateWhenPortBelowRangeReturnsError() {
	s.cfg.Server.Port = invalidPortLow

	err := s.cfg.Validate()

	s.ErrorIs(err, config.ErrPortOutOfRange)
	s.Contains(err.Error(), wantPortMessage)
}

func (s *ValidateTestSuite) TestValidateWhenPortAboveRangeReturnsError() {
	s.cfg.Server.Port = invalidPortHigh

	s.ErrorIs(s.cfg.Validate(), config.ErrPortOutOfRange)
}

func (s *ValidateTestSuite) TestValidateWhenReadTimeoutZeroReturnsError() {
	s.cfg.Server.ReadTimeout = invalidDuration

	s.ErrorIs(s.cfg.Validate(), config.ErrReadTimeoutNotPositive)
}

func (s *ValidateTestSuite) TestValidateWhenWriteTimeoutNegativeReturnsError() {
	s.cfg.Server.WriteTimeout = invalidNegDuration

	s.ErrorIs(s.cfg.Validate(), config.ErrWriteTimeoutNotPositive)
}

func (s *ValidateTestSuite) TestValidateWhenIdleTimeoutZeroReturnsError() {
	s.cfg.Server.IdleTimeout = invalidDuration

	s.ErrorIs(s.cfg.Validate(), config.ErrIdleTimeoutNotPositive)
}

func (s *ValidateTestSuite) TestValidateWhenRequestTimeoutZeroReturnsError() {
	s.cfg.Server.RequestTimeout = invalidDuration

	s.ErrorIs(s.cfg.Validate(), config.ErrRequestTimeoutNotPositive)
}

func (s *ValidateTestSuite) TestValidateWhenRequestTimeoutEqualsWriteTimeoutReturnsError() {
	s.cfg.Server.RequestTimeout = s.cfg.Server.WriteTimeout

	s.ErrorIs(s.cfg.Validate(), config.ErrRequestTimeoutNotBelowWrite)
}

func (s *ValidateTestSuite) TestValidateWhenRequestTimeoutAboveWriteTimeoutReturnsError() {
	s.cfg.Server.RequestTimeout = s.cfg.Server.WriteTimeout + validTimeout

	s.ErrorIs(s.cfg.Validate(), config.ErrRequestTimeoutNotBelowWrite)
}

func (s *ValidateTestSuite) TestValidateWhenRequestTimeoutZeroDoesNotAlsoReportOrdering() {
	s.cfg.Server.RequestTimeout = invalidDuration

	s.NotErrorIs(s.cfg.Validate(), config.ErrRequestTimeoutNotBelowWrite)
}

func (s *ValidateTestSuite) TestValidateWhenShutdownGraceZeroReturnsError() {
	s.cfg.Server.ShutdownGracePeriod = invalidDuration

	s.ErrorIs(s.cfg.Validate(), config.ErrShutdownGraceNotPositive)
}

func (s *ValidateTestSuite) TestValidateWhenShutdownDrainNegativeReturnsError() {
	s.cfg.Server.ShutdownDrainDelay = invalidNegDuration

	s.ErrorIs(s.cfg.Validate(), config.ErrShutdownDrainNegative)
}

func (s *ValidateTestSuite) TestValidateWhenShutdownDrainZeroIsValid() {
	s.cfg.Server.ShutdownDrainDelay = invalidDuration

	s.NoError(s.cfg.Validate())
}

func (s *ValidateTestSuite) TestValidateWhenMaxBodyZeroReturnsError() {
	s.cfg.Server.MaxBodyBytes = invalidMaxBody

	s.ErrorIs(s.cfg.Validate(), config.ErrMaxBodyNotPositive)
}

func (s *ValidateTestSuite) TestValidateWhenNoCORSOriginsReturnsError() {
	s.cfg.CORS.AllowedOrigins = nil

	s.ErrorIs(s.cfg.Validate(), config.ErrCORSOriginsEmpty)
}

func (s *ValidateTestSuite) TestValidateWhenCORSOriginIsWildcardReturnsError() {
	s.cfg.CORS.AllowedOrigins = []string{originWildcard}

	s.ErrorIs(s.cfg.Validate(), config.ErrCORSWildcard)
}

func (s *ValidateTestSuite) TestValidateWhenCORSOriginHasPathReturnsError() {
	s.cfg.CORS.AllowedOrigins = []string{originWithPath}

	s.ErrorIs(s.cfg.Validate(), config.ErrCORSOriginInvalid)
}

func (s *ValidateTestSuite) TestValidateWhenCORSOriginHasNoSchemeReturnsError() {
	s.cfg.CORS.AllowedOrigins = []string{originWithoutScheme}

	s.ErrorIs(s.cfg.Validate(), config.ErrCORSOriginInvalid)
}

func (s *ValidateTestSuite) TestValidateWhenCORSOriginSchemeIsNotHTTPReturnsError() {
	s.cfg.CORS.AllowedOrigins = []string{originWrongScheme}

	s.ErrorIs(s.cfg.Validate(), config.ErrCORSOriginInvalid)
}

func (s *ValidateTestSuite) TestValidateWhenCORSOriginHasQueryReturnsError() {
	s.cfg.CORS.AllowedOrigins = []string{originWithQuery}

	s.ErrorIs(s.cfg.Validate(), config.ErrCORSOriginInvalid)
}

func (s *ValidateTestSuite) TestValidateWhenCORSOriginHasFragmentReturnsError() {
	s.cfg.CORS.AllowedOrigins = []string{originWithFragment}

	s.ErrorIs(s.cfg.Validate(), config.ErrCORSOriginInvalid)
}

func (s *ValidateTestSuite) TestValidateWhenCORSOriginHasCredentialsReturnsError() {
	s.cfg.CORS.AllowedOrigins = []string{originWithCredentials}

	s.ErrorIs(s.cfg.Validate(), config.ErrCORSOriginInvalid)
}

func (s *ValidateTestSuite) TestValidateWhenCORSOriginHasNoHostReturnsError() {
	s.cfg.CORS.AllowedOrigins = []string{originWithoutHost}

	s.ErrorIs(s.cfg.Validate(), config.ErrCORSOriginInvalid)
}

func (s *ValidateTestSuite) TestValidateWhenCORSOriginUnparsableReturnsError() {
	s.cfg.CORS.AllowedOrigins = []string{originUnparsable}

	s.ErrorIs(s.cfg.Validate(), config.ErrCORSOriginInvalid)
}

func (s *ValidateTestSuite) TestValidateWhenOperandLengthBelowRangeReturnsError() {
	s.cfg.Calculator.MaxOperandLength = invalidOperandLow

	err := s.cfg.Validate()

	s.ErrorIs(err, config.ErrOperandLengthOutOfRange)
	s.Contains(err.Error(), wantOperandMessage)
}

func (s *ValidateTestSuite) TestValidateWhenOperandLengthAboveRangeReturnsError() {
	s.cfg.Calculator.MaxOperandLength = invalidOperandHigh

	s.ErrorIs(s.cfg.Validate(), config.ErrOperandLengthOutOfRange)
}

func (s *ValidateTestSuite) TestValidateWhenExponentBelowRangeReturnsError() {
	s.cfg.Calculator.MaxExponent = invalidExponentLow

	s.ErrorIs(s.cfg.Validate(), config.ErrExponentOutOfRange)
}

func (s *ValidateTestSuite) TestValidateWhenExponentAboveRangeReturnsError() {
	s.cfg.Calculator.MaxExponent = invalidExponentHigh

	s.ErrorIs(s.cfg.Validate(), config.ErrExponentOutOfRange)
}

func (s *ValidateTestSuite) TestValidateWhenDecimalPrecisionBelowRangeReturnsError() {
	s.cfg.Calculator.DecimalPrecision = invalidPrecisionLow

	s.ErrorIs(s.cfg.Validate(), config.ErrDecimalPrecisionOutOfRange)
}

func (s *ValidateTestSuite) TestValidateWhenDecimalPrecisionAboveRangeReturnsError() {
	s.cfg.Calculator.DecimalPrecision = invalidPrecisionHigh

	s.ErrorIs(s.cfg.Validate(), config.ErrDecimalPrecisionOutOfRange)
}

func (s *ValidateTestSuite) TestValidateWhenSeveralRulesBrokenReportsEveryOne() {
	s.cfg.Server.Port = invalidPortLow
	s.cfg.Server.ReadTimeout = invalidDuration
	s.cfg.CORS.AllowedOrigins = []string{originWildcard, originWithPath}
	s.cfg.Calculator.MaxExponent = invalidExponentHigh

	err := s.cfg.Validate()

	s.ErrorIs(err, config.ErrPortOutOfRange)
	s.ErrorIs(err, config.ErrReadTimeoutNotPositive)
	s.ErrorIs(err, config.ErrCORSWildcard)
	s.ErrorIs(err, config.ErrCORSOriginInvalid)
	s.ErrorIs(err, config.ErrExponentOutOfRange)
}

func (s *ValidateTestSuite) TestValidateErrorMessageNamesTheOffendingKey() {
	s.cfg.Calculator.MaxExponent = invalidExponentLow

	s.Contains(s.cfg.Validate().Error(), wantExponentMessageFormat)
}
