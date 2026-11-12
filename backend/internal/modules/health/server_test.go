package health_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/contextkeys"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/httpx"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/health"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/health/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/health/mock"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

type ServerTestSuite struct {
	suite.Suite
	ctx      context.Context
	ctrl     *gomock.Controller
	mockCore *mock.MockICore
	handler  *health.HTTPHandler
	router   chi.Router
	recorder *httptest.ResponseRecorder
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

func (s *ServerTestSuite) SetupTest() {
	s.ctx = contextkeys.WithRequestID(context.Background(), testRequestID)
	s.ctrl = gomock.NewController(s.T())
	s.mockCore = mock.NewMockICore(s.ctrl)
	s.recorder = httptest.NewRecorder()
	s.router = chi.NewRouter()
	s.router.MethodNotAllowed(httpx.MethodNotAllowed)
	s.router.NotFound(httpx.NotFound)
	s.handler = health.NewHTTPHandler(s.mockCore)
	s.handler.RegisterRoutes(s.router)
}

func (s *ServerTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *ServerTestSuite) request(method, path string) {
	req := httptest.NewRequest(method, path, nil).WithContext(s.ctx)
	s.router.ServeHTTP(s.recorder, req)
}

// call invokes a handler method directly so the mock sees the exact request
// context; chi would wrap it with routing state.
func (s *ServerTestSuite) call(h http.HandlerFunc) {
	h(s.recorder, httptest.NewRequest(http.MethodGet, entities.PathReady, nil).WithContext(s.ctx))
}

func (s *ServerTestSuite) TestLiveReturns200Ok() {
	s.request(http.MethodGet, entities.PathLive)

	s.Equal(http.StatusOK, s.recorder.Code)
	s.JSONEq(wantLiveBody, s.recorder.Body.String())
	s.Equal(wantContentType, s.recorder.Header().Get(constants.HeaderContentType))
}

func (s *ServerTestSuite) TestLiveNeverConsultsCore() {
	// The strict mock fails the test if Live calls IsReady: liveness must not
	// depend on readiness state.
	s.request(http.MethodGet, entities.PathLive)

	s.Equal(http.StatusOK, s.recorder.Code)
}

func (s *ServerTestSuite) TestReadyReturns200WhenCoreIsReady() {
	s.mockCore.EXPECT().IsReady(gomock.Eq(s.ctx)).Return(true).Times(1)

	s.call(s.handler.Ready)

	s.Equal(http.StatusOK, s.recorder.Code)
	s.JSONEq(wantReadyBody, s.recorder.Body.String())
	s.Equal(wantContentType, s.recorder.Header().Get(constants.HeaderContentType))
}

func (s *ServerTestSuite) TestReadyReturns503ErrorShapeWhenCoreIsNotReady() {
	s.mockCore.EXPECT().IsReady(gomock.Eq(s.ctx)).Return(false).Times(1)

	s.call(s.handler.Ready)

	var got httpx.ErrorResponse
	s.Require().NoError(json.Unmarshal(s.recorder.Body.Bytes(), &got))
	s.Equal(http.StatusServiceUnavailable, s.recorder.Code)
	s.Equal(apperror.CodeServiceUnavailable, got.Error.Code)
	s.Equal(apperror.MsgServiceUnavailable, got.Error.Message)
	s.Equal(testRequestID, got.Error.RequestID)
}

func (s *ServerTestSuite) TestReadyNotReadyBodyHidesInternalCause() {
	s.mockCore.EXPECT().IsReady(gomock.Eq(s.ctx)).Return(false).Times(1)

	s.call(s.handler.Ready)

	s.NotContains(s.recorder.Body.String(), testNotReadyText)
}

func (s *ServerTestSuite) TestReadyIsRoutedAndReceivesRequestContext() {
	hasRequestID := gomock.Cond(func(x any) bool {
		ctx, ok := x.(context.Context)
		if !ok {
			return false
		}
		id, found := contextkeys.RequestID(ctx)
		return found && id == testRequestID
	})
	s.mockCore.EXPECT().IsReady(hasRequestID).Return(true).Times(1)

	s.request(http.MethodGet, entities.PathReady)

	s.Equal(http.StatusOK, s.recorder.Code)
}

func (s *ServerTestSuite) TestPostToLiveIsMethodNotAllowed() {
	s.request(http.MethodPost, entities.PathLive)

	s.Equal(http.StatusMethodNotAllowed, s.recorder.Code)
}

func (s *ServerTestSuite) TestPostToReadyIsMethodNotAllowed() {
	s.request(http.MethodPost, entities.PathReady)

	s.Equal(http.StatusMethodNotAllowed, s.recorder.Code)
}

func (s *ServerTestSuite) TestUnknownHealthPathIsNotFound() {
	s.request(http.MethodGet, testUnknownPath)

	s.Equal(http.StatusNotFound, s.recorder.Code)
}
