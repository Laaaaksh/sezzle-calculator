package calculator

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/config"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/httpx"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// HTTPHandler is the transport layer of the calculator module. It depends on
// ICore, not on Core, so tests can replace the business logic. Every response
// goes through httpx, which sets Content-Type and Cache-Control: no-store.
type HTTPHandler struct {
	core             ICore
	maxOperandLength int
}

// NewHTTPHandler returns a handler backed by core. cfg supplies the operand
// length limit the parser enforces before core sees the input.
func NewHTTPHandler(core ICore, cfg config.CalculatorConfig) *HTTPHandler {
	return &HTTPHandler{core: core, maxOperandLength: cfg.MaxOperandLength}
}

// RegisterRoutes mounts the calculator endpoints on r, the versioned API
// router. Routes are flat so any other method on a path gets the router's JSON
// 405.
func (h *HTTPHandler) RegisterRoutes(r chi.Router) {
	r.Post(entities.PathCalculate, h.Calculate)
	r.Get(entities.PathOperations, h.Operations)
}

// Calculate serves POST /api/v1/calculate. The checks run cheapest first and
// the first failure is returned: media type, body shape, required fields,
// operand syntax and limits, then the core. Core only ever sees decimals.
func (h *HTTPHandler) Calculate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if aerr := checkContentType(r.Header.Get(constants.HeaderContentType)); aerr != nil {
		httpx.WriteError(ctx, w, aerr)
		return
	}
	req, aerr := decodeRequest(r.Body)
	if aerr != nil {
		httpx.WriteError(ctx, w, aerr)
		return
	}
	if problems := invalidFields(req); len(problems) > 0 {
		httpx.WriteError(ctx, w, validationFailed(problems))
		return
	}
	operands, aerr := ParseOperands(req.Operands, h.maxOperandLength)
	if aerr != nil {
		httpx.WriteError(ctx, w, aerr)
		return
	}
	result, aerr := h.core.Calculate(ctx, req.Operation, operands)
	if aerr != nil {
		httpx.WriteError(ctx, w, aerr)
		return
	}
	// The request is echoed as sent: the client sees its own strings, not a
	// re-formatted copy of the parsed decimals.
	httpx.WriteJSON(ctx, w, http.StatusOK, entities.CalculateResponse{
		Operation: req.Operation,
		Operands:  req.Operands,
		Result:    result.Text,
	})
}

// Operations serves GET /api/v1/operations.
func (h *HTTPHandler) Operations(w http.ResponseWriter, r *http.Request) {
	infos := h.core.ListOperations(r.Context())
	if infos == nil {
		// The contract promises an array; nil would encode as null.
		infos = []entities.OperationInfo{}
	}
	httpx.WriteJSON(r.Context(), w, http.StatusOK, entities.OperationsResponse{Operations: infos})
}

// checkContentType accepts application/json, optionally with a utf-8 charset.
// Parsing the header (and not comparing the raw string) is what lets
// "application/json; charset=UTF-8" through.
func checkContentType(header string) apperror.IError {
	mediaType, params, err := mime.ParseMediaType(header)
	if err != nil {
		return unsupportedMediaType(errors.Join(ErrContentTypeMissing, err))
	}
	if mediaType != constants.ContentTypeJSON {
		return unsupportedMediaType(ErrContentTypeNotJSON)
	}
	if charset, ok := params[paramCharset]; ok && !strings.EqualFold(charset, charsetUTF8) {
		return unsupportedMediaType(ErrCharsetNotUTF8)
	}
	return nil
}

func unsupportedMediaType(cause error) apperror.IError {
	return apperror.New(apperror.CodeUnsupportedMediaType, cause)
}

// decodeRequest reads exactly one JSON object. Unknown fields are refused so a
// misspelled field fails loudly instead of being ignored, and anything after
// the object is refused so two concatenated requests are never half-read.
func decodeRequest(body io.Reader) (entities.CalculateRequest, apperror.IError) {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	var req entities.CalculateRequest
	if err := decoder.Decode(&req); err != nil {
		if errors.Is(err, io.EOF) {
			err = ErrEmptyBody
		}
		return entities.CalculateRequest{}, decodeFailure(err)
	}
	var extra json.RawMessage
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return req, nil
	}
	if err == nil {
		err = ErrTrailingData
	}
	return entities.CalculateRequest{}, decodeFailure(err)
}

// decodeFailure maps a read or decode error to its code. A body cut off by the
// size limit is the client sending too much, not malformed JSON, so it is
// PAYLOAD_TOO_LARGE; everything else is BAD_REQUEST.
func decodeFailure(err error) apperror.IError {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		return apperror.New(apperror.CodePayloadTooLarge, err).
			WithField(constants.LogKeyLimit, tooLarge.Limit)
	}
	return apperror.New(apperror.CodeBadRequest, err)
}

// invalidFields names each required field the request leaves out, and an
// operation name over the length bound. An empty operands array is not
// missing: it is present with the wrong count, which the core reports as
// INVALID_OPERAND_COUNT. The length bound runs before the core so a huge name
// never reaches an error field or the logs.
func invalidFields(req entities.CalculateRequest) map[string]string {
	problems := map[string]string{}
	if req.Operation == "" {
		problems[entities.FieldNameOperation] = entities.ProblemRequired
	} else if len(req.Operation) > entities.MaxOperationNameLength {
		problems[entities.FieldNameOperation] = entities.ProblemTooLong
	}
	if req.Operands == nil {
		problems[entities.FieldNameOperands] = entities.ProblemRequired
	}
	return problems
}

// validationFailed builds the error for invalid fields. The problems become
// the error fields, which reach both the client and the log.
func validationFailed(problems map[string]string) apperror.IError {
	aerr := apperror.New(apperror.CodeValidationFailed, ErrMissingRequestField)
	for field, problem := range problems {
		aerr = aerr.WithField(field, problem)
	}
	return aerr
}
