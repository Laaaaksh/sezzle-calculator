package apperror_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/suite"

	// Blank import: the calculator codes register themselves in that package's
	// init, and the docs-sync tests below must see them in apperror.Codes.
	_ "github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// apiDoc is the REST contract. Its error tables must agree with the code, so a
// status changed in one place cannot drift from the other.
var apiDoc = filepath.Join("..", "..", "..", "docs", "API.md")

// rowPattern matches a table row that starts with a code: | `CODE` | 422 |
var rowPattern = regexp.MustCompile("(?m)^\\| `([A-Z_]+)` \\| (\\d{3}) \\|")

type ErrorCodesDocTestSuite struct {
	suite.Suite
	apiStatuses map[apperror.Code]string
}

func TestErrorCodesDocTestSuite(t *testing.T) {
	suite.Run(t, new(ErrorCodesDocTestSuite))
}

func (s *ErrorCodesDocTestSuite) SetupTest() {
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
