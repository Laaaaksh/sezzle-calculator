package calculator_test

const (
	testFakeSymbol      = "?"
	testFakeDescription = "Fake operation for tests."

	testOpFake     = "fake"
	testOpUnary    = "unary"
	testOpMissing  = "missing"
	opModulo       = "modulo"
	testOpBlank    = ""
	testOpBoom     = "boom"
	testFakeResult = "7.500"
	testFakeText   = "7.5"
	testNegZero    = "-0.00"
	testZeroText   = "0"
	testBigText    = "100.00"
	testBigWant    = "100"

	testMaxOperandLength   = 8
	testMaxExponent        = 1000
	testPrecision          = 20
	testDefaultMaxLength   = 64
	testTightExponent      = 5
	testExponentAboveTight = "6"

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

const (
	testRequestID = "req-calc-1"

	testCalcPath       = "/calculate"
	testOperationsPath = "/operations"

	contentTypeJSON       = "application/json"
	contentTypeJSONUTF8   = "application/json; charset=UTF-8"
	contentTypeJSONLatin  = "application/json; charset=iso-8859-1"
	contentTypeMixedCase  = "Application/JSON"
	contentTypePlain      = "text/plain"
	contentTypeUnreadable = ";;;"
	contentTypeAbsent     = ""

	cacheNoStore = "no-store"

	bodyDivide          = `{"operation":"divide","operands":["10","4"]}`
	bodyDivideSpaced    = `{"operation":"divide","operands":["10.0","4.00"]}`
	bodySqrt            = `{"operation":"sqrt","operands":["144"]}`
	bodyEmptyOperands   = `{"operation":"add","operands":[]}`
	bodyEmpty           = ``
	bodyBlank           = " \n\t "
	bodyNotJSON         = `not json`
	bodyCutOff          = `{"operation":"divide","operands":["10",`
	bodyUnknownField    = `{"operation":"add","operands":["1","2"],"extra":true}`
	bodyNumericOperand  = `{"operation":"add","operands":[1,2]}`
	bodyNumericOp       = `{"operation":7,"operands":["1","2"]}`
	bodyArray           = `[1,2]`
	bodyTwoObjects      = bodyDivide + bodyDivide
	bodyTrailingText    = bodyDivide + ` x`
	bodyTrailingSpace   = bodyDivide + " \n"
	bodyNoFields        = `{}`
	bodyNull            = `null`
	bodyNoOperation     = `{"operands":["1","2"]}`
	bodyBlankOperation  = `{"operation":"","operands":["1","2"]}`
	bodyNoOperands      = `{"operation":"add"}`
	bodyNullOperands    = `{"operation":"add","operands":null}`
	bodyBadOperand      = `{"operation":"add","operands":["1","1e5"]}`
	bodyLongOperand     = `{"operation":"add","operands":["1","123456789012345678901234567890123456789012345678901234567890123456"]}`
	bodyOperationHead   = `{"operation":"`
	bodyOperationTail   = `","operands":["1","2"]}`
	bodyAdd             = `{"operation":"add","operands":["1","2"]}`
	bodyTooManyOperands = `{"operation":"add","operands":["1","2","3","4","5","6","7","8","9","10","11","12","13","14","15","16","17"]}`
	bodyThreeOperands   = `{"operation":"add","operands":["1","2","3"]}`
	bodyDuplicateKey    = `{"operation":"add","operands":["1","2"],"operands":["9","9"]}`
	bodyCaseVariantDup  = `{"operation":"add","operands":["1","2"],"OPERANDS":["9","9"]}`
	bodyWrongCaseKey    = `{"Operation":"add","operands":["1","2"]}`
	bodyDuplicateUnk    = `{"operation":"add","operands":["1","2"],"extra":1,"extra":2}`
	bodyUnknownOp       = `{"operation":"modulo","operands":["1","2"]}`

	testLongOperationLen = 33
	testHugeOperationLen = 1 << 20
	testSmallBodyLimit   = 1 << 10
	problemTooLong       = "is too long"

	testHandlerMaxLen       = 64
	testHandlerMaxExponent  = 1000
	testHandlerResultDigits = 640
	testHandlerPrecision    = 20
	testOversizeLimit       = int64(10)
	testTrailingLimit       = int64(60)
	testPadding             = 200

	wantResultText   = "2.5"
	wantSqrtText     = "12"
	wantDivideResp   = `{"operation":"divide","operands":["10","4"],"result":"2.5"}`
	wantSpacedResp   = `{"operation":"divide","operands":["10.0","4.00"],"result":"2.5"}`
	wantSqrtResp     = `{"operation":"sqrt","operands":["144"],"result":"12"}`
	wantOperationsJS = `{"operations":[{"name":"add","arity":2,"symbol":"+","description":"Adds the two operands."},{"name":"sqrt","arity":1,"symbol":"√","description":"Returns the square root of the operand."}],"limits":{"maxOperandLength":64,"maxExponent":1000,"decimalPrecision":20,"maxResultDigits":640}}`
	wantEmptyOpsJS   = `{"operations":[],"limits":{"maxOperandLength":64,"maxExponent":1000,"decimalPrecision":20,"maxResultDigits":640}}`

	fieldOperation = "operation"
	fieldOperands  = "operands"
	fieldKey       = "key"
	reasonDup      = "duplicate"
	reasonCase     = "wrong_case"
	problemText    = "is required"
	problemMany    = "too many"
	wantAddEcho    = `{"operation":"add","operands":["1","2"],"result":"3"}`
	wantAddText    = "3"
	secretCause    = "pq: connection refused to db-7"
)

const (
	testResultDigits      = 640
	testSmallResultDigits = 5
	testMaxProductDigits  = 128
	testBelowProductLimit = 127

	sizeAtLimit        = "12345"
	sizeOverLimit      = "123456"
	sizeNegAtLimit     = "-12345"
	sizeFractionAt     = "0.12345"
	sizeFractionOver   = "0.123456"
	sizeTrailingZeros  = "1.50000000"
	sizeAddLeft        = "999"
	sizeAddRight       = "1"
	sizeSubtractLeft   = "0"
	sizeSubtractRight  = "1000"
	sizeMultiplyLeft   = "100"
	sizeMultiplyRight  = "100"
	sizeDivideLeft     = "1"
	sizeDivideRight    = "3"
	sizePercentLeft    = "50"
	sizePercentRight   = "2000"
	sizePowerBase      = "10"
	sizePowerExponent  = "3"
	sizeSqrtOperand    = "2"
	sizeTinyBudget     = 3
	sizeMaxOperandNine = "9"
	sizeSquareEight    = "8"
	sizeSquareZero     = "0"
	sizeSquareOne      = "1"
)
