package entities

// CalculateRequest is the body of POST /api/v1/calculate. Operands are strings, never JSON numbers: most clients parse a
// number into a float64, which cannot hold 0.1 or a 20-digit integer exactly. A string arrives untouched, so 0.1 + 0.2
// can be exactly 0.3. A nil Operands means the field was absent; an empty slice means the client sent [], which the
// arity check rejects.
type CalculateRequest struct {
	Operation OperationType `json:"operation"`
	Operands  []string      `json:"operands"`
}
