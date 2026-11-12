package entities

// CalculateRequest is the body of POST /api/v1/calculate. Operands are
// strings, never JSON numbers, so no digit is lost before the service sees it
// (see the calculator package comment).
//
// A nil Operands means the field was absent or null; an empty slice means the
// client sent []. The handler treats the first as a missing field and passes
// the second on, where the arity check rejects it.
type CalculateRequest struct {
	Operation OperationType `json:"operation"`
	Operands  []string      `json:"operands"`
}
