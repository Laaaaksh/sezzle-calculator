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
}
