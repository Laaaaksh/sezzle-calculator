package entities

// Limits are the bounds the service enforces on one calculation. They come
// from configuration, are handed to the module as plain values, and are also
// published by GET /api/v1/operations so a client can disable what the server
// would refuse. The JSON names are the wire contract.
type Limits struct {
	// MaxOperandLength bounds the digits of an operand and the length of its text.
	MaxOperandLength int `json:"maxOperandLength"`
	// MaxExponent bounds the magnitude of a power exponent.
	MaxExponent int `json:"maxExponent"`
	// DecimalPrecision is the number of decimal places a division or root keeps.
	DecimalPrecision int `json:"decimalPrecision"`
	// MaxResultDigits bounds the digits of any result, integer and fraction
	// digits together. Every operation is held to it by the core, so one rule
	// covers add, multiply, power and the rest.
	MaxResultDigits int `json:"maxResultDigits"`
}

// ResultDigitsFactor sets the result budget as a multiple of the operand
// length: a result may be ten times longer than any operand the service
// accepts, and no longer.
const ResultDigitsFactor = 10

// ResultDigitBudget derives the result digit limit from the operand length, so
// the two limits move together when the operand length is configured.
func ResultDigitBudget(maxOperandLength int) int {
	return maxOperandLength * ResultDigitsFactor
}
