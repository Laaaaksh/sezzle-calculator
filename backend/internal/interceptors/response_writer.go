package interceptors

import "net/http"

// responseWriter remembers the status and whether the header was sent, which
// the access log needs to report and recovery needs to decide whether an error
// body can still be written.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

// wrapWriter returns w as a *responseWriter, reusing it if an outer
// middleware already wrapped it so all layers see the same state.
func wrapWriter(w http.ResponseWriter) *responseWriter {
	if rw, ok := w.(*responseWriter); ok {
		return rw
	}
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

// WriteHeader records the first status only; later calls are ignored by
// net/http as well.
func (rw *responseWriter) WriteHeader(status int) {
	if !rw.wroteHeader {
		rw.status = status
		rw.wroteHeader = true
	}
	rw.ResponseWriter.WriteHeader(status)
}

// Write marks the header as sent, because net/http sends an implicit 200.
func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.wroteHeader = true
	return rw.ResponseWriter.Write(b)
}

// Unwrap lets http.ResponseController reach the underlying writer for
// flushing and deadlines.
func (rw *responseWriter) Unwrap() http.ResponseWriter { return rw.ResponseWriter }
