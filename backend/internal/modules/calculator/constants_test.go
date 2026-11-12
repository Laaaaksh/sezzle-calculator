package calculator_test

const (
	testOpFake     = "fake"
	testOpUnary    = "unary"
	testOpMissing  = "missing"
	testOpBlank    = ""
	testOpBoom     = "boom"
	testFakeResult = "7.500"
	testFakeText   = "7.5"
	testNegZero    = "-0.00"
	testZeroText   = "0"
	testBigText    = "100.00"
	testBigWant    = "100"

	testMaxOperandLength = 8
	testMaxExponent      = 1000
	testPrecision        = 20
	testDefaultMaxLength = 64

	testOperandOk       = "12345678"
	testOperandTooLong  = "123456789"
	testOperandNegLong  = "-123456789"
	testOperandDecimal  = "1.2345678"
	testOperandLongFrac = "0.123456789"
	testTinyExponent    = "1e-100"
	testHugeExponent    = "1e100"

	testIndexSecond = 1
	testIndexFirst  = 0
)

const (
	testParseMaxLen   = 8
	testParseLongText = "this is not a number but it is far longer than the limit allows"
	testFullwidth     = "１２３"
	testNoLeakText    = "secret99"
	testWordsText     = "abc"
)
