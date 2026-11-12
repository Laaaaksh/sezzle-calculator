package entities

// CalculateResponse is the 200 body of POST /api/v1/calculate. Operation and
// Operands repeat the request exactly as sent, so a client can show the
// expression beside the result without keeping its own state. Result is the
// normalised string from Result.Text.
type CalculateResponse struct {
	Operation OperationType `json:"operation"`
	Operands  []string      `json:"operands"`
	Result    string        `json:"result"`
}

// OperationInfo describes one operation for GET /api/v1/operations.
type OperationInfo struct {
	Name        OperationType `json:"name"`
	Arity       int           `json:"arity"`
	Symbol      string        `json:"symbol"`
	Description string        `json:"description"`
}

// OperationsResponse is the body of GET /api/v1/operations. Operations lets a
// client discover what the server supports; Limits lets it disable input the
// server would refuse. Both sit in an object so a field can be added later
// without breaking clients.
type OperationsResponse struct {
	Operations []OperationInfo `json:"operations"`
	Limits     Limits          `json:"limits"`
}
