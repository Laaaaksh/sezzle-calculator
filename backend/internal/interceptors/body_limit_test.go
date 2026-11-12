package interceptors_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/httpx"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/interceptors"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

type BodyLimitTestSuite struct {
	suite.Suite
	recorder *httptest.ResponseRecorder
	reached  bool
	readBody string
	readErr  error
	handler  http.Handler
}

func TestBodyLimitTestSuite(t *testing.T) {
	suite.Run(t, new(BodyLimitTestSuite))
}

func (s *BodyLimitTestSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
	s.reached, s.readBody, s.readErr = false, "", nil
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		s.reached = true
		data, err := io.ReadAll(r.Body)
		s.readBody, s.readErr = string(data), err
	})
	s.handler = interceptors.BodyLimit(testBodyLimit)(next)
}

func (s *BodyLimitTestSuite) TestDeclaredLengthOverLimitGetsJSON413() {
	req := httptest.NewRequest(http.MethodPost, testPath, strings.NewReader(testOverLimit))

	s.handler.ServeHTTP(s.recorder, req)

	var got httpx.ErrorResponse
	s.Require().NoError(json.Unmarshal(s.recorder.Body.Bytes(), &got))
	s.Equal(http.StatusRequestEntityTooLarge, s.recorder.Code)
	s.Equal(apperror.CodePayloadTooLarge, got.Error.Code)
}

func (s *BodyLimitTestSuite) TestDeclaredLengthOverLimitNeverReachesHandler() {
	req := httptest.NewRequest(http.MethodPost, testPath, strings.NewReader(testOverLimit))

	s.handler.ServeHTTP(s.recorder, req)

	s.False(s.reached)
}

func (s *BodyLimitTestSuite) TestBodyExactlyAtLimitIsReadInFull() {
	body := strings.Repeat("a", int(testBodyLimit))

	s.handler.ServeHTTP(s.recorder, httptest.NewRequest(http.MethodPost, testPath, strings.NewReader(body)))

	s.NoError(s.readErr)
	s.Equal(body, s.readBody)
}

func (s *BodyLimitTestSuite) TestBodyUnderLimitPassesThrough() {
	s.handler.ServeHTTP(s.recorder, httptest.NewRequest(http.MethodPost, testPath, strings.NewReader(testBody)))

	s.True(s.reached)
	s.Equal(testBody, s.readBody)
}

func (s *BodyLimitTestSuite) TestUnknownLengthBodyOverLimitFailsOnRead() {
	req := httptest.NewRequest(http.MethodPost, testPath, strings.NewReader(testOverLimit))
	req.ContentLength = -1

	s.handler.ServeHTTP(s.recorder, req)

	var tooLarge *http.MaxBytesError
	s.True(s.reached)
	s.True(errors.As(s.readErr, &tooLarge))
	s.Equal(testBodyLimit, tooLarge.Limit)
}

func (s *BodyLimitTestSuite) TestRequestWithoutBodyPasses() {
	s.handler.ServeHTTP(s.recorder, httptest.NewRequest(http.MethodGet, testPath, nil))

	s.True(s.reached)
	s.NoError(s.readErr)
}

func (s *BodyLimitTestSuite) TestUnknownLengthOverLimitBehindWrapperClosesConnection() {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		httpx.WriteError(r.Context(), w, apperror.New(apperror.CodePayloadTooLarge, nil))
	})
	client, stop := startPipeServer(interceptors.AccessLog(interceptors.BodyLimit(testBodyLimit)(next)))
	defer stop()
	req, err := http.NewRequest(http.MethodPost, testPipeURL, io.NopCloser(strings.NewReader(testOverLimit)))
	s.Require().NoError(err)
	req.ContentLength = -1

	resp, err := client.Do(req)
	s.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()

	s.Equal(http.StatusRequestEntityTooLarge, resp.StatusCode)
	s.True(resp.Close)
}
