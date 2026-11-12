package boot_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/boot"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/config"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/contextkeys"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/httpx"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator"
	calcentities "github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	calcmock "github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/mock"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/health"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/health/mock"
)

type RouterTestSuite struct {
	suite.Suite
	ctrl     *gomock.Controller
	mockCore *mock.MockICore
	calcCore *calcmock.MockICore
	router   chi.Router
	recorder *httptest.ResponseRecorder
}

func TestRouterTestSuite(t *testing.T) {
	suite.Run(t, new(RouterTestSuite))
}

func (s *RouterTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockCore = mock.NewMockICore(s.ctrl)
	s.calcCore = calcmock.NewMockICore(s.ctrl)
	s.recorder = httptest.NewRecorder()
	cfg := &config.Config{
		Server:     config.ServerConfig{MaxBodyBytes: testBodyLimit, RequestTimeout: testTimeout},
		CORS:       config.CORSConfig{AllowedOrigins: []string{testOrigin}},
		Calculator: config.CalculatorConfig{MaxOperandLength: e2eMaxOperandLength},
	}
	modules := boot.Modules{
		Health: &health.Module{
			Core:    s.mockCore,
			Handler: health.NewHTTPHandler(s.mockCore),
		},
		Calculator: &calculator.Module{
			Core:    s.calcCore,
			Handler: calculator.NewHTTPHandler(s.calcCore, cfg.Calculator),
		},
	}
	ping := func(api chi.Router) {
		api.Get(testPingRoute, func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(testPingBody))
		})
	}
	s.router = boot.NewRouter(cfg, modules, ping)
}

func (s *RouterTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// carriesRequestID matches the request context the router hands a handler: the
// request-ID middleware must have stored a non-empty ID in it.
func (s *RouterTestSuite) carriesRequestID() gomock.Matcher {
	return gomock.Cond(func(x any) bool {
		ctx, ok := x.(context.Context)
		if !ok {
			return false
		}
		id, found := contextkeys.RequestID(ctx)
		return found && id != ""
	})
}

func (s *RouterTestSuite) serve(req *http.Request) {
	s.router.ServeHTTP(s.recorder, req)
}

func (s *RouterTestSuite) errorBody() httpx.ErrorBody {
	var got httpx.ErrorResponse
	s.Require().NoError(json.Unmarshal(s.recorder.Body.Bytes(), &got))
	return got.Error
}

func (s *RouterTestSuite) TestUnknownPathReturnsJSONNotFound() {
	s.serve(httptest.NewRequest(http.MethodGet, testMissingPath, nil))

	body := s.errorBody()
	s.Equal(http.StatusNotFound, s.recorder.Code)
	s.Equal(wantNotFoundCode, string(body.Code))
	s.Equal(wantContentType, s.recorder.Header().Get("Content-Type"))
	s.NotEmpty(body.RequestID)
}

func (s *RouterTestSuite) TestWrongMethodReturnsJSONMethodNotAllowed() {
	s.serve(httptest.NewRequest(http.MethodPost, testLivePath, nil))

	body := s.errorBody()
	s.Equal(http.StatusMethodNotAllowed, s.recorder.Code)
	s.Equal(wantMethodCode, string(body.Code))
	s.NotEmpty(body.RequestID)
	s.Equal(wantLiveAllow, s.recorder.Header().Get(headerAllow))
}

func (s *RouterTestSuite) TestAPIRoutesAreServedUnderVersionedPrefix() {
	s.serve(httptest.NewRequest(http.MethodGet, testAPIPingPath, nil))

	s.Equal(http.StatusOK, s.recorder.Code)
	s.Equal(testPingBody, s.recorder.Body.String())
}

func (s *RouterTestSuite) TestAPIRoutesAreNotServedWithoutVersionedPrefix() {
	s.serve(httptest.NewRequest(http.MethodGet, testPingRoute, nil))

	s.Equal(http.StatusNotFound, s.recorder.Code)
}

func (s *RouterTestSuite) TestWrongMethodUnderAPIPrefixReturnsJSONMethodNotAllowed() {
	s.serve(httptest.NewRequest(http.MethodPost, testAPIPingPath, nil))

	s.Equal(http.StatusMethodNotAllowed, s.recorder.Code)
	s.Equal(wantMethodCode, string(s.errorBody().Code))
	s.Equal(wantPingAllow, s.recorder.Header().Get(headerAllow))
}

func (s *RouterTestSuite) TestUnknownAPIPathReturnsJSONNotFoundFromSubrouter() {
	s.serve(httptest.NewRequest(http.MethodGet, testAPIPath, nil))

	s.Equal(http.StatusNotFound, s.recorder.Code)
	s.Equal(wantNotFoundCode, string(s.errorBody().Code))
}

func (s *RouterTestSuite) TestLiveRouteServesOk() {
	s.serve(httptest.NewRequest(http.MethodGet, testLivePath, nil))

	s.Equal(http.StatusOK, s.recorder.Code)
	s.JSONEq(wantLiveBody, s.recorder.Body.String())
}

func (s *RouterTestSuite) TestReadyRouteServesReadyWhenCoreIsReady() {
	s.mockCore.EXPECT().IsReady(gomock.Cond(func(any) bool { return true })).Return(true).Times(1)

	s.serve(httptest.NewRequest(http.MethodGet, testReadyPath, nil))

	s.Equal(http.StatusOK, s.recorder.Code)
	s.JSONEq(wantReadyBody, s.recorder.Body.String())
}

func (s *RouterTestSuite) TestIncomingRequestIDIsEchoed() {
	req := httptest.NewRequest(http.MethodGet, testLivePath, nil)
	req.Header.Set(wantHeaderTemplate, testRequestID)

	s.serve(req)

	s.Equal(testRequestID, s.recorder.Header().Get(wantHeaderTemplate))
}

func (s *RouterTestSuite) TestGeneratedRequestIDAppearsOnErrorBodyAndHeader() {
	s.serve(httptest.NewRequest(http.MethodGet, testMissingPath, nil))

	header := s.recorder.Header().Get(wantHeaderTemplate)
	s.NotEmpty(header)
	s.Equal(header, s.errorBody().RequestID)
}

func (s *RouterTestSuite) TestCORSHeaderAddedForConfiguredOrigin() {
	req := httptest.NewRequest(http.MethodGet, testLivePath, nil)
	req.Header.Set(headerOrigin, testOrigin)

	s.serve(req)

	s.Equal(testOrigin, s.recorder.Header().Get(headerACAO))
}

func (s *RouterTestSuite) TestCORSHeaderAbsentForOtherOrigin() {
	req := httptest.NewRequest(http.MethodGet, testLivePath, nil)
	req.Header.Set(headerOrigin, testOtherOrigin)

	s.serve(req)

	s.Empty(s.recorder.Header().Get(headerACAO))
}

func (s *RouterTestSuite) TestOversizedBodyIsRejectedBeforeRouting() {
	req := httptest.NewRequest(http.MethodPost, testLivePath, io.NopCloser(strings.NewReader(testOversizedBody)))
	req.ContentLength = int64(len(testOversizedBody))

	s.serve(req)

	s.Equal(http.StatusRequestEntityTooLarge, s.recorder.Code)
	s.Equal(wantTooLargeCode, string(s.errorBody().Code))
}

func (s *RouterTestSuite) TestCalculateRouteIsMountedUnderVersionedPrefix() {
	s.calcCore.EXPECT().
		Calculate(s.carriesRequestID(), gomock.Eq(calcentities.OpAdd), gomock.Eq(testAddOperands())).
		Return(calcentities.Result{Text: testAddResult}, nil).
		Times(1)
	req := httptest.NewRequest(http.MethodPost, testCalculatePath, strings.NewReader(testAddBody))
	req.Header.Set(constants.HeaderContentType, wantContentType)

	s.serve(req)

	s.Equal(http.StatusOK, s.recorder.Code)
	s.JSONEq(wantAddResponse, s.recorder.Body.String())
}

func (s *RouterTestSuite) TestOperationsRouteIsMountedUnderVersionedPrefix() {
	s.calcCore.EXPECT().
		ListOperations(s.carriesRequestID()).
		Return([]calcentities.OperationInfo{}).
		Times(1)

	s.serve(httptest.NewRequest(http.MethodGet, testOperationsPath, nil))

	s.Equal(http.StatusOK, s.recorder.Code)
	s.JSONEq(wantNoOperationsBody, s.recorder.Body.String())
}

func (s *RouterTestSuite) TestCalculateRouteIsNotServedWithoutVersionedPrefix() {
	s.serve(httptest.NewRequest(http.MethodPost, testCalculateBare, nil))

	s.Equal(http.StatusNotFound, s.recorder.Code)
}
