package httprpc

import (
	"bufio"
	"io"
	"net/http"

	"github.com/liuwangchen/toy/transport/encoding"
)

type IClientStream interface {
	// SendMsg is generally called by generated code. On error, SendMsg aborts
	// the stream. If the error was generated by the client, the status is
	// returned directly; otherwise, io.EOF is returned and the status of
	// the stream may be discovered using RecvMsg.
	//
	// SendMsg blocks until:
	//   - There is sufficient flow control to schedule m with the transport, or
	//   - The stream is done, or
	//   - The stream breaks.
	//
	// SendMsg does not wait until the message is received by the server. An
	// untimely stream closure may result in lost messages. To ensure delivery,
	// users should ensure the RPC completed successfully using RecvMsg.
	//
	// It is safe to have a goroutine calling SendMsg and another goroutine
	// calling RecvMsg on the same stream at the same time, but it is not safe
	// to call SendMsg on the same stream in different goroutines. It is also
	// not safe to call CloseSend concurrently with SendMsg.
	SendMsg(m interface{}) error
	// RecvMsg blocks until it receives a message into m or the stream is
	// done. It returns io.EOF when the stream completes successfully. On
	// any other error, the stream is aborted and the error contains the RPC
	// status.
	//
	// It is safe to have a goroutine calling SendMsg and another goroutine
	// calling RecvMsg on the same stream at the same time, but it is not
	// safe to call RecvMsg on the same stream in different goroutines.
	RecvMsg(m interface{}) error
}

type IServerStream interface {
	// SendMsg sends a message. On error, SendMsg aborts the stream and the
	// error is returned directly.
	//
	// SendMsg blocks until:
	//   - There is sufficient flow control to schedule m with the transport, or
	//   - The stream is done, or
	//   - The stream breaks.
	//
	// SendMsg does not wait until the message is received by the client. An
	// untimely stream closure may result in lost messages.
	//
	// It is safe to have a goroutine calling SendMsg and another goroutine
	// calling RecvMsg on the same stream at the same time, but it is not safe
	// to call SendMsg on the same stream in different goroutines.
	//
	// It is not safe to modify the message after calling SendMsg. Tracing
	// libraries and stats handlers may use the message lazily.
	SendMsg(m interface{}) error
	// RecvMsg blocks until it receives a message into m or the stream is
	// done. It returns io.EOF when the client has performed a CloseSend. On
	// any non-EOF error, the stream is aborted and the error contains the
	// RPC status.
	//
	// It is safe to have a goroutine calling SendMsg and another goroutine
	// calling RecvMsg on the same stream at the same time, but it is not
	// safe to call RecvMsg on the same stream in different goroutines.
	RecvMsg(m interface{}) error
}

type HttpServerStream struct {
	w   io.Writer
	enc encoding.Codec
}

func NewHttpServerStream(w io.Writer, codec encoding.Codec) *HttpServerStream {
	return &HttpServerStream{
		w:   w,
		enc: codec,
	}
}

func (this *HttpServerStream) SendMsg(v interface{}) error {
	b, err := this.enc.Marshal(v)
	if err != nil {
		return err
	}
	_, err = this.w.Write(b)
	if err != nil {
		return err
	}
	_, err = this.w.Write([]byte("\n"))
	if err != nil {
		return err
	}
	flusher, ok := this.w.(http.Flusher)
	if ok {
		flusher.Flush()
	}
	return nil
}

func (this *HttpServerStream) RecvMsg(m interface{}) error {
	return nil
}

type HttpClientStream struct {
	r   io.Reader
	dec encoding.Codec
}

func NewHttpClientStream(r io.Reader, codec encoding.Codec) *HttpClientStream {
	return &HttpClientStream{
		r:   bufio.NewReader(r),
		dec: codec,
	}
}

func (this *HttpClientStream) SendMsg(m interface{}) error {
	return nil
}

func (this *HttpClientStream) RecvMsg(v interface{}) error {
	b, err := this.r.(*bufio.Reader).ReadBytes('\n')
	if len(b) > 0 {
		errM := this.dec.Unmarshal(b, v)
		if errM != nil {
			return errM
		}
	}
	if err == io.EOF {
		closer, ok := this.r.(io.Closer)
		if ok {
			closer.Close()
		}
	}
	return err
}
