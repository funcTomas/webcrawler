package module

import (
	"net/http"
	"strings"
	"testing"
)

type testingReader struct {
	sr *strings.Reader
}

func (r testingReader) Read(b []byte) (n int, err error) {
	return r.sr.Read(b)
}

func (r testingReader) Close() error {
	return nil
}

func TestRequest(t *testing.T) {
	method := "GET"
	expectedURLStr := "https://github.com/gopcp"
	expectedHTTPReq, _ := http.NewRequest(method, expectedURLStr, nil)
	expectedDepth := uint32(0)
	req := NewRequest(expectedHTTPReq, expectedDepth)
	if req == nil {
		t.Fatal("Cound not create request")
	}
	if _, ok := interface{}(req).(Data); !ok {
		t.Fatal("Request did not implement Data!")
	}
	expectedValidity := true
	valid := req.Valid()
	if valid != expectedValidity {
		t.Fatalf("Inconsistent validity for request, expected: %v, actual: %v", expectedValidity, valid)
	}
	if req.HTTPReq() != expectedHTTPReq {
		t.Fatalf("Inconsistent HTTP request, expected: %#v, actual: %#v", expectedHTTPReq, req.httpReq)
	}
	if req.Depth() != expectedDepth {
		t.Fatalf("Inconsistent depth for request, expected: %d, actual: %d", expectedDepth, req.Depth())
	}
	expectedHTTPReq.URL = nil
	req = NewRequest(expectedHTTPReq, expectedDepth)
	expectedValidity = false
	valid = req.Valid()
	if valid != expectedValidity {
		t.Fatalf("Inconsistent validity for request, expected: %v, actual: %v", expectedValidity, valid)
	}
	req = NewRequest(nil, expectedDepth)
	expectedValidity = false
	valid = req.Valid()
	if valid != expectedValidity {
		t.Fatalf("Inconsistent validity for request, expected: %v, actual: %v", expectedValidity, valid)
	}
}

func TestResponse(t *testing.T) {
}
