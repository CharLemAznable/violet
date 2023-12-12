package elf

import (
	"bytes"
	"github.com/CharLemAznable/violet/internal/types"
	"io"
	"net/http"
)

func DumpRequestBody(request types.Req) (body []byte, err error) {
	request.Body, body, err = drainBody(request.Body)
	return body, err
}

func DumpResponseBody(response types.Rsp) (body []byte, err error) {
	response.Body, body, err = drainBody(response.Body)
	return body, err
}

func drainBody(body io.ReadCloser) (copy io.ReadCloser, content []byte, err error) {
	if body == nil || body == http.NoBody {
		// No copying needed. Preserve the magic sentinel meaning of NoBody.
		return http.NoBody, nil, nil
	}
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(body); err != nil {
		return nil, nil, err
	}
	if err = body.Close(); err != nil {
		return nil, nil, err
	}
	copy = io.NopCloser(&buf)
	reader := io.NopCloser(bytes.NewReader(buf.Bytes()))
	content, err = io.ReadAll(reader)
	return copy, content, err
}
