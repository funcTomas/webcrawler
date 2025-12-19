package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"webcrawler/module"

	"github.com/PuerkitoBio/goquery"
)

type Item struct {
	Id       string `json:"id"`
	Title    string `json:"title"`
	Url      string `json:"url"`
	Date     string `json:"date"`
	NodeId   string `json:"nodeId"`
	ImgCount string `json:"imgCount"`
}
type News struct {
	Items []Item `json:"items"`
}

func splitFunc(r rune) bool {
	return r == '/' || r == '-' || r == '.'
}

func genResponseParsersV2() []module.ParseResponse {
	parseLink := func(httpResp *http.Response, respDepth uint32) ([]module.Data, []error) {
		if httpResp == nil {
			return nil, []error{fmt.Errorf("nil HTTP response")}
		}
		httpReq := httpResp.Request
		if httpReq == nil {
			return nil, []error{fmt.Errorf("nil HTTP request")}
		}
		reqURL := httpReq.URL
		if httpResp.StatusCode != 200 {
			err := fmt.Errorf("unsupported status code %d (requestURL: %s)", httpResp.StatusCode, reqURL)
			return nil, []error{err}
		}
		if httpResp.Body == nil {
			err := fmt.Errorf("nil HTTP response body (requestURL: %s)", reqURL)
			return nil, []error{err}
		}
		dataList := make([]module.Data, 0)
		var matchedContentType bool
		if httpResp.Header != nil {
			contentTypes := httpResp.Header["Content-Type"]
			for _, ct := range contentTypes {
				if strings.HasPrefix(ct, "application/javascript") {
					matchedContentType = true
					break
				}
			}
		}
		if !matchedContentType {
			return nil, nil
		}
		body, err := io.ReadAll(httpResp.Body)
		if err != nil {
			return nil, []error{err}
		}
		var news News
		if err := json.Unmarshal(body, &news); err != nil {
			return nil, []error{err}
		}
		errs := make([]error, 0)
		for _, v := range news.Items {
			httpReq, err := http.NewRequest("GET", v.Url, nil)
			if err != nil {
				errs = append(errs, err)
			} else {
				req := module.NewRequest(httpReq, respDepth)
				dataList = append(dataList, module.Data(req))
			}
		}
		return dataList, errs
	}
	parseText := func(httpResp *http.Response, respDepth uint32) ([]module.Data, []error) {
		if httpResp == nil {
			return nil, []error{fmt.Errorf("nil HTTP response")}
		}
		httpReq := httpResp.Request
		if httpReq == nil {
			return nil, []error{fmt.Errorf("nil HTTP request")}
		}
		reqURL := httpReq.URL
		if httpResp.StatusCode != 200 {
			err := fmt.Errorf("unsupported status code %d (requestURL: %s)", httpResp.StatusCode, reqURL)
			return nil, []error{err}
		}
		body := httpResp.Body
		if body == nil {
			err := fmt.Errorf("nil HTTP response body (requestURL: %s)", reqURL)
			return nil, []error{err}
		}
		dataList := make([]module.Data, 0)
		var matchedContentType bool
		if httpResp.Header != nil {
			contentTypes := httpResp.Header["Content-Type"]
			for _, ct := range contentTypes {
				if strings.HasPrefix(ct, "text/html") {
					matchedContentType = true
					break
				}
			}
		}
		if !matchedContentType {
			return nil, nil
		}
		doc, err := goquery.NewDocumentFromReader(body)
		if err != nil {
			return nil, []error{err}
		}
		ss := strings.FieldsFunc(reqURL.Path, splitFunc)
		catagory := "default_catagory"
		if len(ss) >= 3 {
			catagory = ss[len(ss)-3]
		}
		title := doc.Find("title").Text()
		var content []string
		doc.Find("div[class=box_pic]~p").Each(func(i int, s *goquery.Selection) {
			content = append(content, s.Text())
		})
		doc.Find("div[class=otitle]~p").Each(func(i int, s *goquery.Selection) {
			content = append(content, s.Text())
		})
		doc.Find("div[class=bza]~p").Each(func(i int, s *goquery.Selection) {
			content = append(content, s.Text())
		})
		doc.Find("div[class=artDet]>p").Each(func(i int, s *goquery.Selection) {
			content = append(content, s.Text())
		})
		item := make(map[string]interface{})
		item["title"] = title
		item["catagory"] = catagory
		item["content"] = strings.Join(content, "\n")
		item["srcUrl"] = reqURL.String()
		dataList = append(dataList, module.Item(item))
		return dataList, nil
	}
	return []module.ParseResponse{parseLink, parseText}

}
