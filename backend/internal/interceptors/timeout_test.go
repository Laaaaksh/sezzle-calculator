package interceptors_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/interceptors"
)

type TimeoutTestSuite struct {
	suite.Suite
	recorder *httptest.ResponseRecorder
	request  *http.Request
}

func TestTimeoutTestSuite(t *testing.T) {
	suite.Run(t, new(TimeoutTestSuite))
}

func (s *TimeoutTestSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
	s.request = httptest.NewRequest(http.MethodGet, testPath, nil)
}

func (s *TimeoutTestSuite) TestContextCarriesDeadlineWithinTimeout() {
	var deadline time.Time
	var hasDeadline bool
	before := time.Now()
	handler := interceptors.Timeout(testTimeout)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		deadline, hasDeadline = r.Context().Deadline()
	}))

	handler.ServeHTTP(s.recorder, s.request)

	s.True(hasDeadline)
	s.False(deadline.Before(before))
	s.False(deadline.After(before.Add(testTimeout + testWaitLimit)))
}

func (s *TimeoutTestSuite) TestHandlerObservesDeadlineExceeded() {
	var ctxErr error
	handler := interceptors.Timeout(testTimeout)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			ctxErr = r.Context().Err()
		case <-time.After(testWaitLimit):
		}
	}))

	handler.ServeHTTP(s.recorder, s.request)

	s.ErrorIs(ctxErr, context.DeadlineExceeded)
}

func (s *TimeoutTestSuite) TestFastHandlerCompletesNormally() {
	handler := interceptors.Timeout(testWaitLimit)(okHandler)

	handler.ServeHTTP(s.recorder, s.request)

	s.Equal(http.StatusOK, s.recorder.Code)
}

func (s *TimeoutTestSuite) TestContextIsReleasedWhenHandlerReturns() {
	var captured context.Context
	handler := interceptors.Timeout(testWaitLimit)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = r.Context()
	}))

	handler.ServeHTTP(s.recorder, s.request)

	s.ErrorIs(captured.Err(), context.Canceled)
}

func (s *TimeoutTestSuite) TestParentCancellationStillPropagates() {
	var ctxErr error
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	handler := interceptors.Timeout(testWaitLimit)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		ctxErr = r.Context().Err()
	}))

	handler.ServeHTTP(s.recorder, s.request.WithContext(ctx))

	s.ErrorIs(ctxErr, context.Canceled)
}
