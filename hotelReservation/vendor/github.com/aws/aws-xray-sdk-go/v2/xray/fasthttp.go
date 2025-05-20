package xray

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/aws/aws-xray-sdk-go/v2/header"
	"github.com/valyala/fasthttp"
)

type FastHTTPHandler interface {
	Handler(SegmentNamer, fasthttp.RequestHandler) fasthttp.RequestHandler
}

type fasthttpHandler struct {
	cfg *Config
}

// NewFastHTTPInstrumentor returns a struct that provides Handle method
// that satisfy fasthttp.RequestHandler interface.
func NewFastHTTPInstrumentor(cfg *Config) FastHTTPHandler {
	return &fasthttpHandler{
		cfg: cfg,
	}
}

// Handler wraps the provided fasthttp.RequestHandler
func (h *fasthttpHandler) Handler(sn SegmentNamer, handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		auxCtx := context.Background()
		if h.cfg != nil {
			auxCtx = context.WithValue(context.Background(), fasthttpContextConfigKey, h.cfg)
			ctx.SetUserValue(fasthttpContextConfigKey, h.cfg)
		}

		name := sn.Name(string(ctx.Request.Host()))
		traceHeader := header.FromString(string(ctx.Request.Header.Peek(TraceIDHeaderKey)))

		req, err := fasthttpToNetHTTPRequest(ctx)
		if err != nil {
			ctx.Logger().Printf("%s", err.Error())
			ctx.Error("Internal Server Error", fasthttp.StatusInternalServerError)
			return
		}

		_, seg := NewSegmentFromHeader(auxCtx, name, req, traceHeader)
		defer seg.Close(nil)

		ctx.SetUserValue(fasthttpContextKey, seg)
		httpCaptureRequest(seg, req)
		fasthttpTrace(seg, handler, ctx, traceHeader)
	}
}

// fasthttpToNetHTTPRequest convert a fasthttp.Request to http.Request
func fasthttpToNetHTTPRequest(ctx *fasthttp.RequestCtx) (*http.Request, error) {
	requestURI := string(ctx.RequestURI())
	rURL, err := url.ParseRequestURI(requestURI)
	if err != nil {
		return nil, fmt.Errorf("cannot parse requestURI %q: %s", requestURI, err)
	}

	req := &http.Request{
		URL:        rURL,
		Host:       string(ctx.Host()),
		RequestURI: requestURI,
		Method:     string(ctx.Method()),
		RemoteAddr: ctx.RemoteAddr().String(),
	}

	hdr := make(http.Header)
	ctx.Request.Header.VisitAll(func(k, v []byte) {
		sk := string(k)
		sv := string(v)
		switch sk {
		case "Transfer-Encoding":
			req.TransferEncoding = append(req.TransferEncoding, sv)
		default:
			hdr.Set(sk, sv)
		}
	})

	req.Header = hdr
	req.TLS = ctx.TLSConnectionState()
	return req, nil
}

func fasthttpTrace(seg *Segment, h fasthttp.RequestHandler, ctx *fasthttp.RequestCtx, traceHeader *header.Header) {
	ctx.Request.Header.Set(TraceIDHeaderKey, generateTraceIDHeaderValue(seg, traceHeader))
	h(ctx)

	seg.Lock()
	seg.GetHTTP().GetResponse().ContentLength = ctx.Response.Header.ContentLength()
	seg.Unlock()
	HttpCaptureResponse(seg, ctx.Response.StatusCode())
}
