package resilience

import (
	"bufio"
	. "github.com/CharLemAznable/violet/internal/types"
	"net/http"
	"strings"
)

func responseFn(responseString string) func(Req) (Rsp, error) {
	if responseString == "" {
		return nil
	}
	return func(req Req) (Rsp, error) {
		return http.ReadResponse(bufio.NewReader(
			strings.NewReader(responseString)), req)
	}
}
