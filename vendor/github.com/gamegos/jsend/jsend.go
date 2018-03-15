/*
Package jsend implements JSend* specification.

You can wrap your http.ResponseWriter:

	jsend.Wrap(w)

Returning object also implements http.ResponseWriter. So you can pass it to your
http middlewares.

Success example:

	jsend.Wrap(w).
		Data(yourData).
		Send()

	// body:
	{
		"status": "success",
		"data": {
			"foo": "bar"
		}
	}

Status field in response body is derived from http status code. Status is "fail"
if code is 4XX, "error" if code is 5XX and "success" otherwise.


Fail:
	jsend.Wrap(w).
		Status(400).
		Data(yourData).
		Send()

	// body:
	{
		"status": "fail",
		"data": {
			"foo": "invalid"
		}
	}

Error:
	jsend.Wrap(w).
		Status(500).
		Message("we are closed").
		Send()

	// body:
	{
		"status": "error",
		"message": "we are closed"
	}


* See http://labs.omniti.com/labs/jsend for jsend spec.
*/
package jsend

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
)

// JSend status codes
const (
	StatusSuccess = "success"
	StatusError   = "error"
	StatusFail    = "fail"
)

const (
	fieldMsg    = "message"
	fieldData   = "data"
	fieldStatus = "status"
)

// Wrap wraps given http.ResponseWriter and returns a response object which
// implements JResponseWriter interface.
//
// If given parameter already implements JResponseWriter "Wrap" returns it
// instead of wrapping it again.
func Wrap(w http.ResponseWriter) JResponseWriter {
	if w, ok := w.(JResponseWriter); ok {
		return w
	}

	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}

	return &Response{rw: w, fields: make(map[string]interface{})}
}

// A JResponseWriter interface extends http.ResponseWriter of go standard library
// to add utility methods for JSend format.
type JResponseWriter interface {
	http.ResponseWriter

	Data(interface{}) JResponseWriter

	Message(string) JResponseWriter

	Status(int) JResponseWriter

	Field(string, interface{}) JResponseWriter

	Send() (int, error)
}

// Response wraps a http.ResponseWriter type and adds jsend methods. Returning
// type implements JResponseWriter which extends http.ResponseWriter.
//
// Response buffers given data and writes nothing until "Send" is called.
type Response struct {
	rw     http.ResponseWriter
	code   int
	sent   bool
	fields map[string]interface{}
	mu     sync.Mutex
}

// Field method allows you to set custom response fields.
func (r *Response) Field(key string, value interface{}) JResponseWriter {
	r.fields[key] = value
	return r
}

// Data sets response's "data" field with given value.
func (r *Response) Data(data interface{}) JResponseWriter {
	return r.Field(fieldData, data)
}

// Message sets response's "message" field with given value.
func (r *Response) Message(msg string) JResponseWriter {
	return r.Field(fieldMsg, msg)
}

// Status sets http statusCode. It is a shorthand for "WriteHeader" method
// in order to keep method chaining.
func (r *Response) Status(code int) JResponseWriter {
	r.code = code
	return r
}

// Header calls Header method of wrapped http.ResponseWriter.
func (r *Response) Header() http.Header {
	return r.rw.Header()
}

// WriteHeader calls WriteHeader method of wrapped http.ResponseWriter.
func (r *Response) WriteHeader(code int) {
	r.code = code
	r.rw.WriteHeader(code)
}

// Write calls Write method of wrapped http.ResponseWriter.
func (r *Response) Write(data []byte) (int, error) {
	return r.rw.Write(data)
}

var errSentAlready = errors.New("jsend: sent already")

// Send encodes and writes buffered data to underlying http response object.
func (r *Response) Send() (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.sent {
		return 0, errSentAlready
	}
	r.sent = true

	if r.code == 0 {
		r.code = 200
	}
	status := getStatus(r.code)

	r.WriteHeader(r.code)
	r.Field(fieldStatus, status)

	if _, hasMsg := r.fields[fieldMsg]; !hasMsg && status == StatusError {
		r.Message(http.StatusText(r.code))
	}

	if _, hasData := r.fields[fieldData]; !hasData && status != StatusError {
		r.Data([]byte(nil))
	}

	j, err := json.Marshal(r.fields)

	if err != nil {
		return 0, err
	}

	return r.Write(j)
}

func getStatus(code int) string {
	switch {
	case code >= 500:
		return StatusError
	case code >= 400 && code < 500:
		return StatusFail
	}

	return StatusSuccess
}
