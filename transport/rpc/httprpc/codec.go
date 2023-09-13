package httprpc

import (
	"io"
	"net/http"

	"github.com/liuwangchen/toy/pkg/httputil"
	"github.com/liuwangchen/toy/transport/encoding"
)

// SupportPackageIsVersion1 These constants should not be referenced from any other code.
const SupportPackageIsVersion1 = true

// DecodeRequestFunc is decode request func.
type DecodeRequestFunc func(*http.Request, interface{}) error

// EncodeResponseFunc is encode response func.
type EncodeResponseFunc func(http.ResponseWriter, *http.Request, interface{}) error

// EncodeErrorFunc is encode error func.
type EncodeErrorFunc func(http.ResponseWriter, *http.Request, error)

// DefaultRequestDecoder decodes the request body to object.
func DefaultRequestDecoder(r *http.Request, v interface{}) error {
	codec := CodecForRequest(r, "Content-Type")
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	if err = codec.Unmarshal(data, v); err != nil {
		return err
	}
	return nil
}

// DefaultResponseEncoder encodes the object to the HTTP response.
func DefaultResponseEncoder(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if v == nil {
		_, err := w.Write(nil)
		return err
	}
	codec := CodecForRequest(r, "Accept")
	data, err := codec.Marshal(v)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", httputil.ContentType(codec.Name()))
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// DefaultErrorEncoder encodes the error to the HTTP response.
func DefaultErrorEncoder(w http.ResponseWriter, r *http.Request, err error) {
	se := map[string]interface{}{"err": err.Error(), "code": 400}
	codec := CodecForRequest(r, "Accept")
	body, err := codec.Marshal(se)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", httputil.ContentType(codec.Name()))
	w.WriteHeader(se["code"].(int))
	_, _ = w.Write(body)
}

// CodecForRequest get encoding.Codec via http.Request
func CodecForRequest(r *http.Request, name string) encoding.Codec {
	for _, accept := range r.Header[name] {
		codec := encoding.GetCodec(httputil.ContentSubtype(accept))
		if codec != nil {
			return codec
		}
	}
	return encoding.GetCodec("json")
}
