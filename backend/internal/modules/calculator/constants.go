package calculator

// Content negotiation for POST /api/v1/calculate. Bodies are JSON in UTF-8;
// a charset parameter is allowed only when it says so, because the JSON
// decoder reads bytes as UTF-8 and would misread any other encoding.
const (
	paramCharset = "charset"
	charsetUTF8  = "utf-8"
)
