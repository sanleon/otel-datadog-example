package handler

import (
	"context"
	"io"
	"net/http"
)

var _ io.ReadCloser = &bodyWrapper{}

type bodyWrapper struct {
	io.ReadCloser
	err error
}

func (w *bodyWrapper) Read(b []byte) (int, error) {
	n, err := w.ReadCloser.Read(b)
	w.err = err
	return n, err
}

func (w *bodyWrapper) Close() error {
	return w.ReadCloser.Close()
}

var _ http.ResponseWriter = &respWriterWrapper{}

// Using this struct for get response information.
type respWriterWrapper struct {
	http.ResponseWriter
	ctx         context.Context
	statusCode  int
	err         error
	wroteHeader bool
}

func (w *respWriterWrapper) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *respWriterWrapper) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
		w.wroteHeader = true
	}
	n, err := w.ResponseWriter.Write(p)
	w.err = err
	return n, err
}

func (w *respWriterWrapper) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
