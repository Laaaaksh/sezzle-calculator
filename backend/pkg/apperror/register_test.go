package apperror_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

const (
	testRegisteredCode    = apperror.Code("TEST_REGISTERED")
	testRegisteredMessage = "A feature code was registered."
	testStatusTooLow      = http.StatusOK
	testStatusTooHigh     = 600
)

type RegisterTestSuite struct {
	suite.Suite
}

func TestRegisterTestSuite(t *testing.T) {
	suite.Run(t, new(RegisterTestSuite))
}

func (s *RegisterTestSuite) TearDownTest() {
	apperror.Unregister(testRegisteredCode)
}

func (s *RegisterTestSuite) TestRegisteredCodeUsesItsStatus() {
	apperror.Register(testRegisteredCode, http.StatusConflict, testRegisteredMessage)

	s.Equal(http.StatusConflict, testRegisteredCode.HTTPStatus())
}

func (s *RegisterTestSuite) TestRegisteredCodeUsesItsDefaultMessage() {
	apperror.Register(testRegisteredCode, http.StatusConflict, testRegisteredMessage)

	s.Equal(testRegisteredMessage, apperror.New(testRegisteredCode, nil).PublicMessage())
}

func (s *RegisterTestSuite) TestRegisteredCodeIsListedByCodes() {
	apperror.Register(testRegisteredCode, http.StatusConflict, testRegisteredMessage)

	s.Contains(apperror.Codes(), testRegisteredCode)
}

func (s *RegisterTestSuite) TestRegisteringTheSameCodeTwicePanics() {
	apperror.Register(testRegisteredCode, http.StatusConflict, testRegisteredMessage)

	s.Panics(func() { apperror.Register(testRegisteredCode, http.StatusConflict, testRegisteredMessage) })
}

func (s *RegisterTestSuite) TestRegisteringAGenericCodePanics() {
	s.Panics(func() { apperror.Register(apperror.CodeNotFound, http.StatusNotFound, testRegisteredMessage) })
}

func (s *RegisterTestSuite) TestRegisteringAnEmptyCodePanics() {
	s.Panics(func() { apperror.Register(apperror.Code(""), http.StatusConflict, testRegisteredMessage) })
}

func (s *RegisterTestSuite) TestRegisteringASuccessStatusPanics() {
	s.Panics(func() { apperror.Register(testRegisteredCode, testStatusTooLow, testRegisteredMessage) })
}

func (s *RegisterTestSuite) TestRegisteringAStatusAbove599Panics() {
	s.Panics(func() { apperror.Register(testRegisteredCode, testStatusTooHigh, testRegisteredMessage) })
}

func (s *RegisterTestSuite) TestAFailedRegistrationLeavesTheCodeUnknown() {
	s.Panics(func() { apperror.Register(testRegisteredCode, testStatusTooLow, testRegisteredMessage) })

	s.NotContains(apperror.Codes(), testRegisteredCode)
}
