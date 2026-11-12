package calculator

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// requestKeys are the only spellings of the request fields the API accepts.
var requestKeys = []string{entities.FieldNameOperation, entities.FieldNameOperands}

// checkKeys rejects a top-level key that is repeated, or that matches a field
// name in a different case, with BAD_REQUEST naming the key. encoding/json
// matches struct fields case-insensitively and keeps the last of two
// duplicates, so {"operands":["1","2"],"OPERANDS":["9","9"]} would silently
// compute on the second list. Exact lowercase names are the only valid
// spelling ("Operation" is refused, not accepted): one spelling per field
// leaves no ambiguity for a proxy or a log reader to resolve differently.
//
// Anything it cannot read as an object is left to the main decode, which
// reports the syntax or shape error; unknown keys likewise (the decoder
// refuses them). Only the keys of the top-level object are checked.
func checkKeys(raw []byte) apperror.IError {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if first, err := decoder.Token(); err != nil || first != json.Delim('{') {
		return nil
	}
	seen := map[string]bool{}
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil
		}
		key, ok := token.(string)
		if !ok {
			return nil
		}
		if aerr := checkKey(key, seen); aerr != nil {
			return aerr
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil
		}
	}
	return nil
}

// checkKey judges one key. Only keys that are, ignoring case, a request field
// are judged here, so the key echoed in the error is as short as a field name.
func checkKey(key string, seen map[string]bool) apperror.IError {
	for _, valid := range requestKeys {
		if !strings.EqualFold(key, valid) {
			continue
		}
		if key != valid {
			return badKey(ErrKeyWrongCase, key, entities.ReasonKeyWrongCase)
		}
		if seen[key] {
			return badKey(ErrDuplicateKey, key, entities.ReasonDuplicateKey)
		}
		seen[key] = true
	}
	return nil
}

func badKey(cause error, key, reason string) apperror.IError {
	return apperror.New(apperror.CodeBadRequest, cause).
		WithField(entities.FieldKey, key).
		WithField(entities.FieldReason, reason)
}
