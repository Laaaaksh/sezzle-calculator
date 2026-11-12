package apperror_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// errorCodesDoc is the repo-relative path of the table clients and the README reuse.
var errorCodesDoc = filepath.Join("..", "..", "..", "docs", "ERROR_CODES.md")

// apiDoc is the REST contract. Its error table must agree with the code, so a
// status changed in one place cannot drift from the other.
var apiDoc = filepath.Join("..", "..", "..", "docs", "API.md")

// rowPattern matches a table row that starts with a code: | `CODE` | 422 |
var rowPattern = regexp.MustCompile("(?m)^\\| `([A-Z_]+)` \\| (\\d{3}) \\|")

type ErrorCodesDocTestSuite struct {
	suite.Suite
	statusByCode map[apperror.Code]string
	apiStatuses  map[apperror.Code]string
}

func TestErrorCodesDocTestSuite(t *testing.T) {
	suite.Run(t, new(ErrorCodesDocTestSuite))
}

func (s *ErrorCodesDocTestSuite) SetupTest() {
	raw, err := os.ReadFile(errorCodesDoc)
	s.Require().NoError(err)
	s.statusByCode = tableRows(raw)
	api, err := os.ReadFile(apiDoc)
	s.Require().NoError(err)
	s.apiStatuses = tableRows(api)
}

func tableRows(raw []byte) map[apperror.Code]string {
	rows := map[apperror.Code]string{}
	for _, match := range rowPattern.FindAllStringSubmatch(string(raw), -1) {
		rows[apperror.Code(match[1])] = match[2]
	}
	return rows
}

func (s *ErrorCodesDocTestSuite) TestEveryCodeIsDocumentedWithItsHTTPStatus() {
	for _, code := range apperror.Codes() {
		s.Equal(fmt.Sprint(code.HTTPStatus()), s.statusByCode[code], "code %s", code)
	}
}

func (s *ErrorCodesDocTestSuite) TestEveryDocumentedCodeExists() {
	known := map[apperror.Code]bool{}
	for _, code := range apperror.Codes() {
		known[code] = true
	}

	for code := range s.statusByCode {
		s.True(known[code], "documented code %s is not defined", code)
	}
}

func (s *ErrorCodesDocTestSuite) TestAPIDocListsEveryCodeWithItsHTTPStatus() {
	for _, code := range apperror.Codes() {
		s.Equal(fmt.Sprint(code.HTTPStatus()), s.apiStatuses[code], "code %s in docs/API.md", code)
	}
}

func (s *ErrorCodesDocTestSuite) TestAPIDocMentionsOnlyCodesThatExist() {
	known := map[apperror.Code]bool{}
	for _, code := range apperror.Codes() {
		known[code] = true
	}

	for code := range s.apiStatuses {
		s.True(known[code], "docs/API.md lists code %s that is not defined", code)
	}
}
