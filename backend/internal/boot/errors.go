package boot

import "errors"

// Stage sentinels let main choose an exit code without parsing text: a bad
// configuration must not be restart-looped, a runtime failure may be.
var (
	ErrConfigLoad = errors.New("boot: load configuration")
	ErrLoggerInit = errors.New("boot: initialise logger")
)

// stageError tags a startup failure with the stage that produced it. Its text
// is the bare cause, because the caller prints its own stage-specific prefix.
type stageError struct {
	stage error
	cause error
}

func (e *stageError) Error() string { return e.cause.Error() }

func (e *stageError) Unwrap() []error { return []error{e.stage, e.cause} }
