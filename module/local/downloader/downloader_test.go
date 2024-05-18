package downloader

import (
	"bufio"
	"net/http"
	"testing"
	"webcrawler/module"
	"webcrawler/module/stub"
)

func TestNew(t *testing.T) {
	mid := module.MID("D1|127.0.0.1:8080")
	httpClient := &http.Client{}
	d, err := New(mid, httpClient, nil)
	if err != nil {
		t.Fatalf("An error occurs when creating a downloader: %s (mid: %s, httpClient: %#v)", err, mid, httpClient)
	}
	if d == nil {
		t.Fatal("Could not create a downloader")
	}
	if d.ID() != mid {
		t.Fatalf("Inconsistent MID for downloader, expected: %s, actual: %s", mid, d.ID())
	}
	mid = module.MID("D127.0.0.1")
	_, err = New(mid, httpClient, nil)
	if err == nil {
		t.Fatalf("No error when creating a downloader with illegal MID %q", mid)
	}
	mid = module.MID("D1|127.0.0.1:8888")
	httpClient = nil
	_, err = New(mid, httpClient, nil)
	if err == nil {
		t.Fatalf("No error when creating a downloader with nil client %q", mid)
	}
}

func TestDownload(t *testing.T) {
	mid := module.MID("D1|127.0.0.1:8080")
	httpClient := &http.Client{}
	d, _ := New(mid, httpClient, nil)
	url := "http://www.baidu.com/robots.txt"
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("An error occurs when creating a HTTP request: %s (url: %s)", err, url)
	}
	depth := uint32(0)
	req := module.NewRequest(httpReq, depth)
	resp, err := d.Download(req)
	if err != nil {
		t.Fatalf("An error occurs when downloading content: %s (req: %#v)", err, req)
	}
	if resp == nil {
		t.Fatalf("response is nil")
	}
	if resp.Depth() != depth {
		t.Fatalf("Inconsitent depth, expected: %d, actual: %d", depth, resp.Depth())
	}
	httpResp := resp.HTTPResp()
	if httpResp == nil {
		t.Fatalf("Invalid HTTP response (url: %s)", url)
	}
	body := httpResp.Body
	if body == nil {
		t.Fatalf("Invalid HTTP response body (url: %s)", url)
	}
	r := bufio.NewReader(body)
	line, _, err := r.ReadLine()
	if err != nil {
		t.Fatalf("An error occurs when reading HTTP response body: %s (url: %s)", err, url)
	}
	lineStr := string(line)
	expectedFirstLine := "User-agent: Baiduspider"
	if lineStr != expectedFirstLine {
		t.Fatalf("Inconsistent first line of the HTTP response body, expected: %s, actual: %s (url: %s)",
			expectedFirstLine, lineStr, url)
	}
	_, err = d.Download(nil)
	if err == nil {
		t.Fatal("No error when downloading nil request")
	}
	url = "http:///www.baidu.com/robots.txt"
	httpReq, err = http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("An error occurs when creating a HTTP request %s (url: %s)", err, url)
	}
	req = module.NewRequest(httpReq, 0)
	_, err = d.Download(req)
	if err == nil {
		t.Fatalf("No error when downloading with invalid url %q", url)
	}
	req = module.NewRequest(nil, 0)
	_, err = d.Download(req)
	if err == nil {
		t.Fatal("No error when downloading with nil HTTP request")
	}
}

func TestCount(t *testing.T) {
	mid := module.MID("D1|127.0.0.1:8080")
	httpClient := &http.Client{}
	d, _ := New(mid, httpClient, nil)
	di := d.(stub.ModuleInternal)
	if di.CalledCount() != 0 {
		t.Fatalf("Inconsitent called count for internal module, expected: %d, actual: %d", 0, di.CalledCount())
	}
	if di.AcceptedCount() != 0 {
		t.Fatalf("Inconsitent accepted count for internal module, expected: %d, actual: %d", 0, di.AcceptedCount())
	}
	if di.CompletedCount() != 0 {
		t.Fatalf("Inconsitent completed count for internal module, expected: %d, actual: %d", 0, di.CompletedCount())
	}
	if di.HandlingNumber() != 0 {
		t.Fatalf("Inconsitent handling number for internal module, expected: %d, actual: %d", 0, di.HandlingNumber())
	}

	d, _ = New(mid, httpClient, nil)
	di = d.(stub.ModuleInternal)
	d.Download(nil)
	if di.CalledCount() != 1 {
		t.Fatalf("Inconsitent called count for internal module, expected: %d, actual: %d", 1, di.CalledCount())
	}
	if di.AcceptedCount() != 0 {
		t.Fatalf("Inconsitent accepted count for internal module, expected: %d, actual: %d", 0, di.AcceptedCount())
	}
	if di.CompletedCount() != 0 {
		t.Fatalf("Inconsitent completed count for internal module, expected: %d, actual: %d", 0, di.CompletedCount())
	}
	if di.HandlingNumber() != 0 {
		t.Fatalf("Inconsitent handling number for internal module, expected: %d, actual: %d", 0, di.HandlingNumber())
	}

	d, _ = New(mid, httpClient, nil)
	di = d.(stub.ModuleInternal)
	url := "http://www.baidu.com/robots.txt"
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("An error when creating a HTTP request %s (url: %s)", err, url)
	}
	req := module.NewRequest(httpReq, 0)
	d.Download(req)
	if di.CalledCount() != 1 {
		t.Fatalf("Inconsitent called count for internal module, expected: %d, actual: %d", 1, di.CalledCount())
	}
	if di.AcceptedCount() != 1 {
		t.Fatalf("Inconsitent accepted count for internal module, expected: %d, actual: %d", 1, di.AcceptedCount())
	}
	if di.CompletedCount() != 1 {
		t.Fatalf("Inconsitent completed count for internal module, expected: %d, actual: %d", 1, di.CompletedCount())
	}
	if di.HandlingNumber() != 0 {
		t.Fatalf("Inconsitent handling number for internal module, expected: %d, actual: %d", 0, di.HandlingNumber())
	}

}
