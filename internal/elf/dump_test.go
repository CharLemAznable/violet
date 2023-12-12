package elf_test

import (
	"bufio"
	"github.com/CharLemAznable/violet/internal/elf"
	"net/http"
	"strings"
	"testing"
)

func TestDump(t *testing.T) {
	req, _ := http.NewRequest("Get", "/", nil)
	requestBody, _ := elf.DumpRequestBody(req)
	if requestBody != nil {
		t.Errorf("Expected requestBody is nil, but got '%s'", string(requestBody))
	}
	requestBody, _ = elf.DumpRequestBody(req)
	if requestBody != nil {
		t.Errorf("Re-Dump Expected requestBody is nil, but got '%s'", string(requestBody))
	}
	rsp, _ := http.ReadResponse(bufio.NewReader(
		strings.NewReader("HTTP/1.1 200 OK\r\n\r\nOK")), req)
	responseBody, _ := elf.DumpResponseBody(rsp)
	if "OK" != string(responseBody) {
		t.Errorf("Expected responseBody is 'OK', but got '%s'", string(responseBody))
	}
	responseBody, _ = elf.DumpResponseBody(rsp)
	if "OK" != string(responseBody) {
		t.Errorf("Re-Dump Expected responseBody is 'OK', but got '%s'", string(responseBody))
	}
}
