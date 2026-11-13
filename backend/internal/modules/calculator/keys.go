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

// checkKeys rejects a repeated top-level key, or one matching a field name in a different case, with BAD_REQUEST.
// encoding/json matches fields case-insensitively and keeps the last duplicate, so {"operands":[..],"OPERANDS":[..]}
// would silently compute on the second list. One exact spelling per field leaves nothing for a proxy or log reader to
// resolve differently. Non-objects and unknown keys are left to the main decode, which reports them.
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
