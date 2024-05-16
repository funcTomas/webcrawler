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
	method := "GET"
	expectedURLStr := "https://github.com/gopcp"
	httpReq, _ := http.NewRequest(method, expectedURLStr, nil)
	expectedHTTPResp := &http.Response{
		Request: httpReq,
		Body: testingReader{
			strings.NewReader("Test Response"),
		},
	}
	expectedDepth := uint32(0)
	resp := NewResponse(expectedHTTPResp, uint32(expectedDepth))
	if resp == nil {
		t.Fatal("Could not create response")
	}
	if _, ok := interface{}(resp).(Data); !ok {
		t.Fatalf("Response did not implement Data!")
	}
	expectedValidity := true
	valid := resp.Valid()
	if valid != expectedValidity {
		t.Fatalf("Inconsistent validity for response, expected: %v, acutal: %v", expectedValidity, valid)
	}
	if resp.HTTPResp() != expectedHTTPResp {
		t.Fatalf("Inconsistent HTTP response for response, expected: %#v, actual: %#v", expectedHTTPResp, resp.HTTPResp())
	}
	if resp.Depth() != expectedDepth {
		t.Fatalf("Inconsistent depth for response, expected: %d, acutal: %d", expectedDepth, resp.Depth())
	}
	expectedHTTPResp.Body = nil
	resp = NewResponse(expectedHTTPResp, expectedDepth)
	expectedValidity = false
	valid = resp.Valid()
	if valid != expectedValidity {
		t.Fatalf("Inconsistent valid for response, expected: %v, actual: %v", expectedValidity, valid)
	}
	resp = NewResponse(nil, expectedDepth)
	expectedValidity = false
	valid = resp.Valid()
	if valid != expectedValidity {
		t.Fatalf("Inconsistent valid for response, expected: %v, actual: %v", expectedValidity, valid)
	}
}

func TestItem(t *testing.T) {
	item := Item(map[string]interface{}{})
	if _, ok := interface{}(item).(Data); !ok {
		t.Fatal("Item did not implement Data!")
	}
	expectedValidity := true
	valid := item.Valid()
	if valid != expectedValidity {
		t.Fatalf("Inconsistent validity for item, expected: %v, actual: %v", expectedValidity, valid)
	}
	item = Item(nil)
	expectedValidity = false
	valid = item.Valid()
	if valid != expectedValidity {
		t.Fatalf("Inconsistent validity for item, expected: %v, acutal: %v", expectedValidity, valid)
	}
}
