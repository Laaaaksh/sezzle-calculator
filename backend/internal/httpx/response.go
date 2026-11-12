// Package httpx writes every JSON response the service sends, so the success
// and error wire formats are defined in exactly one place.
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
	// Fields maps an input field to a client-safe problem description. It is
	// reserved for validation errors and omitted when empty.
	Fields map[string]string `json:"fields,omitempty"`
}

// fallbackBody is sent when a response cannot be encoded. It is a constant so
// the failure path itself cannot fail.
const fallbackBody = `{"error":{"code":"` + string(apperror.CodeInternal) +
	`","message":"` + apperror.MsgInternal + `"}}`

// WriteJSON writes v as JSON with status. The body is encoded before anything
// is sent, so an unencodable value becomes a clean 500 instead of a truncated
// 200. ctx is the request context, so an encode failure is logged with the
// request ID.
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

// WriteError logs err in full (code, cause, fields) and sends only its public
// part, so internals such as driver errors never reach a client.
func WriteError(ctx context.Context, w http.ResponseWriter, err apperror.IError) {
	WriteErrorWithFields(ctx, w, err, nil)
}

// WriteErrorWithFields is WriteError plus a map of client-safe problems per
// input field. The caller builds the map from text it controls (never from the
// cause or from user input), because every value in it reaches the client.
func WriteErrorWithFields(ctx context.Context, w http.ResponseWriter, err apperror.IError, fields map[string]string) {
	logError(ctx, err)
	public := apperror.ToPublic(err)
	body := ErrorBody{Code: public.Code, Message: public.Message, Fields: fields}
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
