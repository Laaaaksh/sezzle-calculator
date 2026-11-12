package operations_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/operations"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

type namedOperation struct {
	name entities.OperationType
}

func (n namedOperation) Name() entities.OperationType { return n.name }

func (n namedOperation) Arity() int { return testFakeArity }

func (n namedOperation) Symbol() string { return testFakeSymbol }

func (n namedOperation) Description() string { return testFakeDescription }

func (n namedOperation) Apply(_ context.Context, _ []decimal.Decimal) (decimal.Decimal, apperror.IError) {
	return decimal.Zero, nil
}

type RegistryTestSuite struct {
	suite.Suite
}

func TestRegistryTestSuite(t *testing.T) {
	suite.Run(t, new(RegistryTestSuite))
}

func (s *RegistryTestSuite) TestLookupReturnsTheRegisteredOperation() {
	one := namedOperation{name: testNameOne}
	registry := operations.NewRegistry(one, namedOperation{name: testNameTwo})

	op, err := registry.Lookup(testNameOne)

	s.Require().Nil(err)
	s.Equal(one, op)
}

func (s *RegistryTestSuite) TestLookupFindsEveryRegisteredOperation() {
	registry := operations.NewRegistry(namedOperation{name: testNameOne}, namedOperation{name: testNameTwo})

	op, err := registry.Lookup(testNameTwo)

	s.Require().Nil(err)
	s.Equal(entities.OperationType(testNameTwo), op.Name())
}

func (s *RegistryTestSuite) TestLookupWhenUnregisteredReturnsUnknownOperation() {
	registry := operations.NewRegistry(namedOperation{name: testNameOne})

	op, err := registry.Lookup(testNameMissing)

	s.Nil(op)
	s.Require().NotNil(err)
	s.Equal(entities.CodeUnknownOperation, err.Code())
	s.ErrorIs(err, operations.ErrUnknownOperation)
	s.Equal(testNameMissing, err.Fields()[entities.FieldOperation])
}

func (s *RegistryTestSuite) TestLookupWhenNameIsAtTheLimitStillEchoesIt() {
	name := strings.Repeat("a", entities.MaxOperationNameLength)

	_, err := operations.NewRegistry().Lookup(entities.OperationType(name))

	s.Require().NotNil(err)
	s.Equal(name, err.Fields()[entities.FieldOperation])
}

func (s *RegistryTestSuite) TestLookupWhenNameIsOverTheLimitOmitsTheOperationField() {
	name := strings.Repeat("a", entities.MaxOperationNameLength+1)

	_, err := operations.NewRegistry().Lookup(entities.OperationType(name))

	s.Require().NotNil(err)
	s.Equal(entities.CodeUnknownOperation, err.Code())
	s.NotContains(err.Fields(), entities.FieldOperation)
}

func (s *RegistryTestSuite) TestLookupWhenNameIsEmptyReturnsUnknownOperation() {
	registry := operations.NewRegistry(namedOperation{name: testNameOne})

	_, err := registry.Lookup(testNameBlank)

	s.Require().NotNil(err)
	s.Equal(entities.CodeUnknownOperation, err.Code())
}

func (s *RegistryTestSuite) TestLookupOnEmptyRegistryReturnsUnknownOperation() {
	_, err := operations.NewRegistry().Lookup(testNameOne)

	s.Require().NotNil(err)
	s.Equal(entities.CodeUnknownOperation, err.Code())
}

func (s *RegistryTestSuite) TestNewRegistryPanicsOnDuplicateName() {
	s.Panics(func() {
		operations.NewRegistry(namedOperation{name: testNameOne}, namedOperation{name: testNameOne})
	})
}

func (s *RegistryTestSuite) TestUnknownOperationIsA422() {
	_, err := operations.NewRegistry().Lookup(testNameOne)

	s.Require().NotNil(err)
	s.Equal(http.StatusUnprocessableEntity, err.HTTPStatus())
}

func (s *RegistryTestSuite) TestListReturnsOperationsInRegistrationOrder() {
	registry := operations.NewRegistry(namedOperation{name: testNameTwo}, namedOperation{name: testNameOne})

	listed := registry.List()

	s.Require().Len(listed, 2)
	s.Equal(entities.OperationType(testNameTwo), listed[0].Name())
	s.Equal(entities.OperationType(testNameOne), listed[1].Name())
}

func (s *RegistryTestSuite) TestListOnEmptyRegistryIsEmpty() {
	s.Empty(operations.NewRegistry().List())
}

func (s *RegistryTestSuite) TestListReturnsACopyThatCannotReorderTheRegistry() {
	registry := operations.NewRegistry(namedOperation{name: testNameOne}, namedOperation{name: testNameTwo})
	registry.List()[0] = namedOperation{name: testNameMissing}

	s.Equal(entities.OperationType(testNameOne), registry.List()[0].Name())
}
