package logger_test

import (
	"context"
	"errors"
	"io"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/contextkeys"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/logger"
)

const (
	testRequestID    = "req-456"
	testUnknownEnv   = "staging"
	testPortValue    = 8080
	testFieldMissing = "field absent"
	testWarnMessage  = "stack trace probe"
	// testStackMarker is a frame present in every stack trace taken inside a test
	// (the caller field alone names only the file).
	testStackMarker = "testing.tRunner"
)

var errTestSync = errors.New("disk on fire")

// errnoSyncer is a write syncer whose Sync returns a fixed error.
type errnoSyncer struct{ err error }

func (e errnoSyncer) Write(p []byte) (int, error) { return len(p), nil }
func (e errnoSyncer) Sync() error                 { return e.err }

type LoggerTestSuite struct {
	suite.Suite
	ctx      context.Context
	observed *observer.ObservedLogs
	restore  func()
}

func TestLoggerTestSuite(t *testing.T) {
	suite.Run(t, new(LoggerTestSuite))
}

func (s *LoggerTestSuite) SetupTest() {
	s.ctx = context.Background()
	core, logs := observer.New(zapcore.DebugLevel)
	s.observed = logs
	s.restore = logger.SetForTest(zap.New(core).Sugar())
}

func (s *LoggerTestSuite) TearDownTest() {
	s.restore()
}

func (s *LoggerTestSuite) TestCtxWithRequestIDAttachesField() {
	ctx := contextkeys.WithRequestID(s.ctx, testRequestID)

	logger.Ctx(ctx).Infow(constants.LogMsgConfigLoaded)

	entries := s.observed.All()
	s.Require().Len(entries, 1)
	s.Equal(constants.LogMsgConfigLoaded, entries[0].Message)
	s.Equal(testRequestID, entries[0].ContextMap()[constants.LogKeyRequestID])
}

func (s *LoggerTestSuite) TestCtxWithoutRequestIDOmitsField() {
	logger.Ctx(s.ctx).Infow(constants.LogMsgConfigLoaded, constants.LogKeyPort, testPortValue)

	entries := s.observed.All()
	s.Require().Len(entries, 1)
	_, present := entries[0].ContextMap()[constants.LogKeyRequestID]
	s.False(present, testFieldMissing)
	s.EqualValues(testPortValue, entries[0].ContextMap()[constants.LogKeyPort])
}

func (s *LoggerTestSuite) TestCtxWithEmptyRequestIDOmitsField() {
	ctx := contextkeys.WithRequestID(s.ctx, "")

	logger.Ctx(ctx).Infow(constants.LogMsgConfigLoaded)

	entries := s.observed.All()
	s.Require().Len(entries, 1)
	_, present := entries[0].ContextMap()[constants.LogKeyRequestID]
	s.False(present, testFieldMissing)
}

func (s *LoggerTestSuite) TestCtxWithNilContextReturnsBaseLogger() {
	//nolint:staticcheck // a nil context must not panic the logger
	logger.Ctx(nil).Infow(constants.LogMsgConfigLoaded)

	s.Len(s.observed.All(), 1)
}

func (s *LoggerTestSuite) TestSyncIgnoresNotATerminalError() {
	restore := logger.SetForTest(syncErrLogger(syscall.ENOTTY))
	defer restore()

	s.NoError(logger.Sync())
}

func (s *LoggerTestSuite) TestSyncIgnoresInvalidArgumentError() {
	restore := logger.SetForTest(syncErrLogger(syscall.EINVAL))
	defer restore()

	s.NoError(logger.Sync())
}

func (s *LoggerTestSuite) TestSyncIgnoresBadFileDescriptorError() {
	restore := logger.SetForTest(syncErrLogger(syscall.EBADF))
	defer restore()

	s.NoError(logger.Sync())
}

func (s *LoggerTestSuite) TestSyncReturnsOtherErrors() {
	restore := logger.SetForTest(syncErrLogger(errTestSync))
	defer restore()

	s.ErrorIs(logger.Sync(), errTestSync)
}

func syncErrLogger(err error) *zap.SugaredLogger {
	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(enc, errnoSyncer{err: err}, zapcore.DebugLevel)
	return zap.New(core).Sugar()
}

// InitializeTestSuite starts from a clean slate: no base logger installed.
type InitializeTestSuite struct {
	suite.Suite
	ctx     context.Context
	restore func()
}

func TestInitializeTestSuite(t *testing.T) {
	suite.Run(t, new(InitializeTestSuite))
}

func (s *InitializeTestSuite) SetupTest() {
	s.ctx = context.Background()
	// Swapping in nil resets the package to its "before Initialize" state.
	s.restore = logger.SetForTest(nil)
}

func (s *InitializeTestSuite) TearDownTest() {
	s.restore()
}

func (s *InitializeTestSuite) TestCtxBeforeInitializeDoesNotPanic() {
	s.NotPanics(func() {
		logger.Ctx(s.ctx).Infow(constants.LogMsgConfigLoaded)
	})
}

func (s *InitializeTestSuite) TestCtxBeforeInitializeNeverReturnsNil() {
	s.NotNil(logger.Ctx(s.ctx))
}

func (s *InitializeTestSuite) TestSyncBeforeInitializeReturnsNil() {
	s.NoError(logger.Sync())
}

func (s *InitializeTestSuite) TestInitializeProdSucceedsAndHidesDebug() {
	s.Require().NoError(logger.Initialize(constants.EnvProd))

	core := logger.Ctx(s.ctx).Desugar().Core()
	s.True(core.Enabled(zapcore.InfoLevel))
	s.False(core.Enabled(zapcore.DebugLevel))
}

func (s *InitializeTestSuite) TestInitializeTestEnvSucceedsAndHidesDebug() {
	s.Require().NoError(logger.Initialize(constants.EnvTest))

	core := logger.Ctx(s.ctx).Desugar().Core()
	s.True(core.Enabled(zapcore.InfoLevel))
	s.False(core.Enabled(zapcore.DebugLevel))
}

func (s *InitializeTestSuite) TestInitializeDevSucceedsAndShowsDebug() {
	s.Require().NoError(logger.Initialize(constants.EnvDev))

	core := logger.Ctx(s.ctx).Desugar().Core()
	s.True(core.Enabled(zapcore.DebugLevel))
}

func (s *InitializeTestSuite) TestInitializeUnknownEnvReturnsError() {
	err := logger.Initialize(testUnknownEnv)

	s.ErrorIs(err, logger.ErrUnknownEnvironment)
	s.Contains(err.Error(), testUnknownEnv)
}

func (s *InitializeTestSuite) TestInitializeUnknownEnvKeepsPreviousLogger() {
	s.Require().NoError(logger.Initialize(constants.EnvDev))

	s.Error(logger.Initialize(testUnknownEnv))

	s.True(logger.Ctx(s.ctx).Desugar().Core().Enabled(zapcore.DebugLevel))
}

// captureStderr runs fn with os.Stderr redirected to a pipe and returns what
// was written. zap resolves "stderr" at build time, so Initialize must be
// called inside fn.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	original := os.Stderr
	os.Stderr = writer
	fn()
	os.Stderr = original
	require.NoError(t, writer.Close())
	out, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())
	return string(out)
}

// Stack traces are reserved for Error and above: a Warn per client 404 with a
// trace attached drowns real failures (zap's dev mode defaults to Warn).
func (s *InitializeTestSuite) TestDevWarnCarriesNoStackTrace() {
	out := captureStderr(s.T(), func() {
		s.Require().NoError(logger.Initialize(constants.EnvDev))
		logger.Ctx(s.ctx).Warnw(testWarnMessage)
	})

	s.Contains(out, testWarnMessage)
	s.NotContains(out, testStackMarker)
}

func (s *InitializeTestSuite) TestDevErrorCarriesStackTrace() {
	out := captureStderr(s.T(), func() {
		s.Require().NoError(logger.Initialize(constants.EnvDev))
		logger.Ctx(s.ctx).Errorw(testWarnMessage)
	})

	s.Contains(out, testStackMarker)
}
