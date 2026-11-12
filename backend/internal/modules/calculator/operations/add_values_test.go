package operations_test

// Operand and result texts used by the add tests, named by value so an
// assertion reads the same as the number it checks. A Long constant holds more
// digits than a readable name can carry.
const (
	addLong1      = "100000000000000000000000000000000"
	addLong2      = "99999999999999999999999999999999"
	addLong3      = "0.000000000000000000000000002"
	addLong4      = "0.000000000000000000000000001"
	addLong5      = "0.123456789012345679"
	addLong6      = "0.1234567890123456789"
	addLong7      = "0.0000000000000000001"
	addNum0       = "0"
	addNum0Pt1    = "0.1"
	addNum0Pt2    = "0.2"
	addNum0Pt3    = "0.3"
	addNum1       = "1"
	addNum2       = "2"
	addNum3       = "3"
	addNum4Pt2    = "4.2"
	addNum5       = "5"
	addNum7Pt5    = "7.5"
	addNumNeg0    = "-0"
	addNumNeg1    = "-1"
	addNumNeg2    = "-2"
	addNumNeg3    = "-3"
	addNumNeg4Pt2 = "-4.2"
	addNumNeg5    = "-5"
)
