package operations_test

const (
	testNameOne     = "one"
	testNameTwo     = "two"
	testNameMissing = "missing"
	testNameBlank   = ""
	testFakeArity   = 2

	testFakeSymbol      = "?"
	testFakeDescription = "Fake operation for tests."
)

const (
	testTwoToTheThousandPrefix = "10715086071862673209484250490600018105614048117055336074437503883703510511249361224931983788156958581275946729175531468251871452856923140435984577574698574803934567774824230985421074605062371141877954182153046474983581941267398767559165543946077062914571196477686542167660429831652624386837205668069376"
	testMaxOperandLength       = 64
	testMaxExponent            = 1000
	testResultBudget           = 640
	testSmallOperandLen        = 2
	testSmallBudget            = 20
	testPowerPrecision         = 20
	testExponentIndex          = 1
	testSqrtIterations         = 1

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

// Expected display metadata, written out so a change to a symbol or a description is a deliberate edit here too.
const (
	wantAddSymbol             = "+"
	wantAddDescription        = "Adds the two operands."
	wantSubtractSymbol        = "-"
	wantSubtractDescription   = "Subtracts the second operand from the first."
	wantMultiplySymbol        = "*"
	wantMultiplyDescription   = "Multiplies the two operands."
	wantDivideSymbol          = "/"
	wantDivideDescription     = "Divides the first operand by the second."
	wantPowerSymbol           = "^"
	wantPowerDescription      = "Raises the first operand to the whole-number power given by the second."
	wantSqrtSymbol            = "\u221a"
	wantSqrtDescription       = "Returns the square root of the operand."
	wantPercentageSymbol      = "%"
	wantPercentageDescription = "Returns the first operand percent of the second."
)

const (
	testTenToTheSixHundredLen   = 601
	testTenToTheLastAllowed     = "639"
	testTenToTheFirstRefused    = "640"
	testRoundBaseZeros          = 62
	testMaxPowerExponent        = "1000"
	testNegMaxPowerExponent     = "-1000"
	testSixHundred              = "600"
	testHalfToTheFiveHundred    = "500"
	testPointOneToTheSixHundred = 602
	testFullOperandDigits       = 64
	testNarrowOperandDigits     = 63
	testNarrowFractionZeros     = 61
	testSignedBoundaryBase      = "-10"
	testSignedBoundaryExponent  = "640"
)

const (
	digitNine  = "9"
	digitZero  = "0"
	digitOne   = "1"
	digitEight = "8"
)

const (
	sizeZero               = "0"
	sizeThousand           = "1000"
	sizeNegThousand        = "-1000"
	sizeTenThousand        = "10000"
	sizeTwelvePointFive    = "12.5"
	sizePointZeroZeroOne   = "0.001"
	sizeOnePointFiveZeros  = "1.500000"
	sizeFiveFractionDigits = "0.12345"
	sizeLongDigits         = 300
	sizeFourDigits         = 4
)
