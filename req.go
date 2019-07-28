package gout

import (
	"bytes"
	"fmt"
	"github.com/guonaihong/gout/decode"
	"github.com/guonaihong/gout/encode"
	"net/http"
	"strings"
)

type Req struct {
	method string
	url    string

	// http body
	bodyEncoder Encoder
	bodyDecoder Decoder

	// http header
	headerEncode interface{}
	headerDecode interface{}

	// query
	queryEncode interface{}

	httpCode *int
	g        *Gout
}

func (r *Req) Reset() {
	r.bodyEncoder = nil
	r.bodyDecoder = nil
	r.httpCode = nil
	r.headerDecode = nil
	r.headerEncode = nil
	r.queryEncode = nil
}

func (r *Req) Do() (err error) {
	b := &bytes.Buffer{}

	defer r.Reset()
	// set http body
	if r.bodyEncoder != nil {
		if err := r.bodyEncoder.Encode(b); err != nil {
			return err
		}
	}

	// set query header
	if r.queryEncode != nil {
		q := encode.NewQueryEncode(nil)
		err = encode.Encode(r.queryEncode, q)
		if err != nil {
			return err
		}

		if query := q.End(); len(query) > 0 {
			r.url += "?" + query
		}
	}

	req, err := http.NewRequest(r.method, r.url, b)
	if err != nil {
		return err
	}

	// set http header
	if r.headerEncode != nil {
		err = encode.Encode(r.headerEncode, encode.NewHeaderEnocde(req))
		if err != nil {
			return err
		}
	}

	resp, err := r.g.Client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if r.headerDecode != nil {
		err = decode.Header.Decode(resp, r.headerDecode)
		if err != nil {
			return err
		}
	}

	if r.bodyDecoder != nil {
		if err := r.bodyDecoder.Decode(resp.Body); err != nil {
			return err
		}
	}

	if r.httpCode != nil {
		*r.httpCode = resp.StatusCode
	}

	return nil
}

func modifyUrl(url string) string {
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		return url
	}

	if strings.HasPrefix(url, ":") {
		return fmt.Sprintf("http://127.0.0.1%s", url)
	}

	if strings.HasPrefix(url, "/") {
		return fmt.Sprintf("http://127.0.0.1%s", url)
	}

	return fmt.Sprintf("http://%s", url)
}

func NewReq(method string, url string, g *Gout) *Req {
	return &Req{method: method, url: modifyUrl(url), g: g}
}
