package xray

import (
	"io"
	"net/http"
)

type responseCapturer struct {
	http.ResponseWriter
	status int
	length int
}

func (w *responseCapturer) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseCapturer) Write(data []byte) (int, error) {
	w.length += len(data)
	return w.ResponseWriter.Write(data)
}

// Returns a wrapped http.ResponseWriter that implements the same optional interfaces
// that the underlying ResponseWriter has.
// Handle every possible combination so that code that checks for the existence of each
// optional interface functions properly.
// Based on https://github.com/felixge/httpsnoop/blob/eadd4fad6aac69ae62379194fe0219f3dbc80fd3/wrap_generated_gteq_1.8.go#L66
func (w *responseCapturer) wrappedResponseWriter() http.ResponseWriter {
	closeNotifier, isCloseNotifier := w.ResponseWriter.(http.CloseNotifier)
	flush, isFlusher := w.ResponseWriter.(http.Flusher)
	hijack, isHijacker := w.ResponseWriter.(http.Hijacker)
	push, isPusher := w.ResponseWriter.(http.Pusher)
	readFrom, isReaderFrom := w.ResponseWriter.(io.ReaderFrom)

	switch {
	case !isCloseNotifier && !isFlusher && !isHijacker && !isPusher && !isReaderFrom:
		return struct {
			http.ResponseWriter
		}{w}

	case isCloseNotifier && !isFlusher && !isHijacker && !isPusher && !isReaderFrom:
		return struct {
			http.ResponseWriter
			http.CloseNotifier
		}{w, closeNotifier}

	case !isCloseNotifier && isFlusher && !isHijacker && !isPusher && !isReaderFrom:
		return struct {
			http.ResponseWriter
			http.Flusher
		}{w, flush}

	case !isCloseNotifier && !isFlusher && isHijacker && !isPusher && !isReaderFrom:
		return struct {
			http.ResponseWriter
			http.Hijacker
		}{w, hijack}

	case !isCloseNotifier && !isFlusher && !isHijacker && isPusher && !isReaderFrom:
		return struct {
			http.ResponseWriter
			http.Pusher
		}{w, push}

	case !isCloseNotifier && !isFlusher && !isHijacker && !isPusher && isReaderFrom:
		return struct {
			http.ResponseWriter
			io.ReaderFrom
		}{w, readFrom}

	case isCloseNotifier && isFlusher && !isHijacker && !isPusher && !isReaderFrom:
		return struct {
			http.ResponseWriter
			http.CloseNotifier
			http.Flusher
		}{w, closeNotifier, flush}

	case isCloseNotifier && !isFlusher && isHijacker && !isPusher && !isReaderFrom:
		return struct {
			http.ResponseWriter
			http.CloseNotifier
			http.Hijacker
		}{w, closeNotifier, hijack}

	case isCloseNotifier && !isFlusher && !isHijacker && isPusher && !isReaderFrom:
		return struct {
			http.ResponseWriter
			http.CloseNotifier
			http.Pusher
		}{w, closeNotifier, push}

	case isCloseNotifier && !isFlusher && !isHijacker && !isPusher && isReaderFrom:
		return struct {
			http.ResponseWriter
			http.CloseNotifier
			io.ReaderFrom
		}{w, closeNotifier, readFrom}

	case !isCloseNotifier && isFlusher && isHijacker && !isPusher && !isReaderFrom:
		return struct {
			http.ResponseWriter
			http.Flusher
			http.Hijacker
		}{w, flush, hijack}

	case !isCloseNotifier && isFlusher && !isHijacker && isPusher && !isReaderFrom:
		return struct {
			http.ResponseWriter
			http.Flusher
			http.Pusher
		}{w, flush, push}

	case !isCloseNotifier && isFlusher && !isHijacker && !isPusher && isReaderFrom:
		return struct {
			http.ResponseWriter
			http.Flusher
			io.ReaderFrom
		}{w, flush, readFrom}

	case !isCloseNotifier && !isFlusher && isHijacker && isPusher && !isReaderFrom:
		return struct {
			http.ResponseWriter
			http.Hijacker
			http.Pusher
		}{w, hijack, push}

	case !isCloseNotifier && !isFlusher && isHijacker && !isPusher && isReaderFrom:
		return struct {
			http.ResponseWriter
			http.Hijacker
			io.ReaderFrom
		}{w, hijack, readFrom}

	case !isCloseNotifier && !isFlusher && !isHijacker && isPusher && isReaderFrom:
		return struct {
			http.ResponseWriter
			http.Pusher
			io.ReaderFrom
		}{w, push, readFrom}

	case isCloseNotifier && isFlusher && isHijacker && !isPusher && !isReaderFrom:
		return struct {
			http.ResponseWriter
			http.CloseNotifier
			http.Flusher
			http.Hijacker
		}{w, closeNotifier, flush, hijack}

	case isCloseNotifier && isFlusher && !isHijacker && isPusher && !isReaderFrom:
		return struct {
			http.ResponseWriter
			http.CloseNotifier
			http.Flusher
			http.Pusher
		}{w, closeNotifier, flush, push}

	case isCloseNotifier && isFlusher && !isHijacker && !isPusher && isReaderFrom:
		return struct {
			http.ResponseWriter
			http.CloseNotifier
			http.Flusher
			io.ReaderFrom
		}{w, closeNotifier, flush, readFrom}

	case isCloseNotifier && !isFlusher && isHijacker && isPusher && !isReaderFrom:
		return struct {
			http.ResponseWriter
			http.CloseNotifier
			http.Hijacker
			http.Pusher
		}{w, closeNotifier, hijack, push}

	case isCloseNotifier && !isFlusher && isHijacker && !isPusher && isReaderFrom:
		return struct {
			http.ResponseWriter
			http.CloseNotifier
			http.Hijacker
			io.ReaderFrom
		}{w, closeNotifier, hijack, readFrom}

	case isCloseNotifier && !isFlusher && !isHijacker && isPusher && isReaderFrom:
		return struct {
			http.ResponseWriter
			http.CloseNotifier
			http.Pusher
			io.ReaderFrom
		}{w, closeNotifier, push, readFrom}

	case !isCloseNotifier && isFlusher && isHijacker && isPusher && !isReaderFrom:
		return struct {
			http.ResponseWriter
			http.Flusher
			http.Hijacker
			http.Pusher
		}{w, flush, hijack, push}

	case !isCloseNotifier && isFlusher && isHijacker && !isPusher && isReaderFrom:
		return struct {
			http.ResponseWriter
			http.Flusher
			http.Hijacker
			io.ReaderFrom
		}{w, flush, hijack, readFrom}

	case !isCloseNotifier && isFlusher && !isHijacker && isPusher && isReaderFrom:
		return struct {
			http.ResponseWriter
			http.Flusher
			http.Pusher
			io.ReaderFrom
		}{w, flush, push, readFrom}

	case !isCloseNotifier && !isFlusher && isHijacker && isPusher && isReaderFrom:
		return struct {
			http.ResponseWriter
			http.Hijacker
			http.Pusher
			io.ReaderFrom
		}{w, hijack, push, readFrom}

	case isCloseNotifier && isFlusher && isHijacker && isPusher && !isReaderFrom:
		return struct {
			http.ResponseWriter
			http.CloseNotifier
			http.Flusher
			http.Hijacker
			http.Pusher
		}{w, closeNotifier, flush, hijack, push}

	case isCloseNotifier && isFlusher && isHijacker && !isPusher && isReaderFrom:
		return struct {
			http.ResponseWriter
			http.CloseNotifier
			http.Flusher
			http.Hijacker
			io.ReaderFrom
		}{w, closeNotifier, flush, hijack, readFrom}

	case isCloseNotifier && isFlusher && !isHijacker && isPusher && isReaderFrom:
		return struct {
			http.ResponseWriter
			http.CloseNotifier
			http.Flusher
			http.Pusher
			io.ReaderFrom
		}{w, closeNotifier, flush, push, readFrom}

	case isCloseNotifier && !isFlusher && isHijacker && isPusher && isReaderFrom:
		return struct {
			http.ResponseWriter
			http.CloseNotifier
			http.Hijacker
			http.Pusher
			io.ReaderFrom
		}{w, closeNotifier, hijack, push, readFrom}

	case !isCloseNotifier && isFlusher && isHijacker && isPusher && isReaderFrom:
		return struct {
			http.ResponseWriter
			http.Flusher
			http.Hijacker
			http.Pusher
			io.ReaderFrom
		}{w, flush, hijack, push, readFrom}

	case isCloseNotifier && isFlusher && isHijacker && isPusher && isReaderFrom:
		return struct {
			http.ResponseWriter
			http.CloseNotifier
			http.Flusher
			http.Hijacker
			http.Pusher
			io.ReaderFrom
		}{w, closeNotifier, flush, hijack, push, readFrom}

	default:
		return struct {
			http.ResponseWriter
		}{w}
	}
}
