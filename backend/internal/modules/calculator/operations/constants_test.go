package operations_test

const (
	testNameOne     = "one"
	testNameTwo     = "two"
	testNameMissing = "missing"
	testNameBlank   = ""
	testFakeArity   = 2
)

const (
	testMaxOperandLength = 64
	testMaxExponent      = 1000
	testResultBudget     = 640
	testSmallOperandLen  = 2
	testSmallBudget      = 20
	testPowerPrecision   = 20
	testExponentIndex    = 1
	testSqrtIterations   = 1

	testSqrtTwoAt20       = "1.4142135623730950488"
	testSqrtTwoAt30       = "1.41421356237309504880168872421"
	testSqrtThreeAt20     = "1.73205080756887729353"
	testSqrtTinyNonSquare = "0.00000000000000141421"
	testBigSquare         = "15241578753238836750495351562536198787501905199875019052100"
	testBigSquareRoot     = "123456789012345678901234567890"
	testTwoToTheHundred   = "1267650600228229401496703205376"
	testTenToTheThirty    = "1000000000000000000000000000000"
	testNearTieSquare     = "6.250000000005000000000001"
	testPercentSquare     = "99999999999999999999"
	testPercentSquareWant = "99999999999999999998000000000000000000.01"
	testTinyUnit          = "0.00000000000000000001"
	testTinyUnitSquareWnt = "0.000000000000000000000000000000000000000001"
)
