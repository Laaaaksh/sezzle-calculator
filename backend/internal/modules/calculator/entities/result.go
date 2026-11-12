package entities

// Result is a normalised calculation outcome. Text is the canonical string the
// API returns: no trailing zeros, no exponent notation and never "-0".
type Result struct {
	Text string
}
