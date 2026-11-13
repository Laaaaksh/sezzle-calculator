// Package httpx writes every JSON response, so the success and error wire formats live in one place.
package httpx

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/contextkeys"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/logger"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// ErrorResponse is the body of every non-2xx response:
// {"error":{"code":"NOT_FOUND","message":"...","requestId":"...","fields":{...}}}.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody is the client-visible error. It carries only public data: the
// cause and internal fields of an apperror.IError are logged, never sent.
type ErrorBody struct {
	Code      apperror.Code `json:"code"`
	Message   string        `json:"message"`
	RequestID string        `json:"requestId,omitempty"`
	// Fields is client-safe context about the failure, such as which operand
	// failed and why. It is omitted when empty.
	Fields map[string]any `json:"fields,omitempty"`
}

// fallbackBody is sent when a response cannot be encoded. It is a constant so
// the failure path itself cannot fail.
const fallbackBody = `{"error":{"code":"` + string(apperror.CodeInternal) +
	`","message":"` + apperror.MsgInternal + `"}}`

// WriteJSON encodes v before sending anything, so an unencodable value becomes a clean 500, not a truncated 200.
// ctx carries the request ID for the failure log.
func WriteJSON(ctx context.Context, w http.ResponseWriter, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		logger.Ctx(ctx).Errorw(constants.LogMsgResponseEncode,
			constants.LogKeyError, err.Error())
		status = http.StatusInternalServerError
		body = []byte(fallbackBody)
	}
	h := w.Header()
	h.Set(constants.HeaderContentType, constants.ContentTypeJSON)
	// Responses are per-request data; an intermediary must never reuse them.
	h.Set(constants.HeaderCacheControl, constants.CacheControlNoStore)
	w.WriteHeader(status)
	// A failed write means the client went away; there is nobody to tell.
	_, _ = w.Write(body)
}

// WriteError logs err in full and sends only the code, public message and fields. Fields hold client-safe context built
// by the code that raised the error; anything internal stays in the cause, which is never sent.
func WriteError(ctx context.Context, w http.ResponseWriter, err apperror.IError) {
	logError(ctx, err)
	public := apperror.ToPublic(err)
	body := ErrorBody{Code: public.Code, Message: public.Message, Fields: err.Fields()}
	if id, ok := contextkeys.RequestID(ctx); ok {
		body.RequestID = id
	}
	WriteJSON(ctx, w, err.HTTPStatus(), ErrorResponse{Error: body})
}

// logError records server faults at error level and client faults at warn, so
// alerting on error-level logs is not triggered by someone mistyping a URL.
// SERVICE_UNAVAILABLE is a deliberate refusal (for example readiness during a
// planned drain), not a fault, so it is logged at warn as well.
func logError(ctx context.Context, err apperror.IError) {
	log := logger.Ctx(ctx)
	args := []any{
		constants.LogKeyCode, string(err.Code()),
		constants.LogKeyStatus, err.HTTPStatus(),
		constants.LogKeyCause, err.Error(),
		constants.LogKeyFields, err.Fields(),
	}
	if isServerFault(err) {
		log.Errorw(constants.LogMsgErrorResponse, args...)
		return
	}
	log.Warnw(constants.LogMsgErrorResponse, args...)
}

func isServerFault(err apperror.IError) bool {
	return err.HTTPStatus() >= http.StatusInternalServerError &&
		err.Code() != apperror.CodeServiceUnavailable
}
