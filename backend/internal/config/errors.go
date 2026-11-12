package config

import (
	"errors"
	"fmt"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
)

// ErrUnknownEnvironment is returned when APP_ENV is not dev, test or prod.
var ErrUnknownEnvironment = errors.New(constants.ErrMsgUnknownEnv)

// Validation errors are package-level values so callers and tests can match
// them with errors.Is instead of comparing message text.
var (
	ErrPortOutOfRange = fmt.Errorf(constants.ErrFmtRange,
		constants.KeyServerPort, constants.PortMin, constants.PortMax)
	ErrReadTimeoutNotPositive = fmt.Errorf(constants.ErrFmtPositive,
		constants.KeyServerReadTimeout)
	ErrWriteTimeoutNotPositive = fmt.Errorf(constants.ErrFmtPositive,
		constants.KeyServerWriteTimeout)
	ErrIdleTimeoutNotPositive = fmt.Errorf(constants.ErrFmtPositive,
		constants.KeyServerIdleTimeout)
	ErrRequestTimeoutNotPositive = fmt.Errorf(constants.ErrFmtPositive,
		constants.KeyServerRequestTimeout)
	// A request deadline at or past the write timeout would let the server cut
	// the connection before the handler can answer with its JSON timeout error.
	ErrRequestTimeoutNotBelowWrite = fmt.Errorf(constants.ErrFmtBelow,
		constants.KeyServerRequestTimeout, constants.KeyServerWriteTimeout)
	ErrShutdownGraceNotPositive = fmt.Errorf(constants.ErrFmtPositive,
		constants.KeyServerShutdownGracePeriod)
	ErrShutdownDrainNegative = fmt.Errorf(constants.ErrFmtNotNeg,
		constants.KeyServerShutdownDrainDelay)
	ErrMaxBodyNotPositive = fmt.Errorf(constants.ErrFmtPositive,
		constants.KeyServerMaxBodyBytes)

	ErrCORSOriginsEmpty = fmt.Errorf(constants.ErrFmtRequired,
		constants.KeyCORSAllowedOrigins)
	ErrCORSWildcard = fmt.Errorf(constants.ErrFmtWildcard,
		constants.KeyCORSAllowedOrigins)
	ErrCORSOriginInvalid = fmt.Errorf(constants.ErrFmtOrigin,
		constants.KeyCORSAllowedOrigins)

	ErrOperandLengthOutOfRange = fmt.Errorf(constants.ErrFmtRange,
		constants.KeyCalculatorMaxOperandLen,
		constants.OperandLengthMin, constants.OperandLengthMax)
	ErrExponentOutOfRange = fmt.Errorf(constants.ErrFmtRange,
		constants.KeyCalculatorMaxExponent,
		constants.ExponentMin, constants.ExponentMax)
	ErrDecimalPrecisionOutOfRange = fmt.Errorf(constants.ErrFmtRange,
		constants.KeyCalculatorDecimalPrec,
		constants.DecimalPrecisionMin, constants.DecimalPrecisionMax)
)
