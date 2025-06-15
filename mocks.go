package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

// curl -X GET \
//   -H "Content-Type: application/x-www-form-encoded" \
//   -H "Authorization: Bearer ${VK_API_TOKEN}" \
//   -o "data/${GROUP_DOMAIN}.json" \
//   "https://api.vk.com/method/wall.get?v=5.199&domain=${GROUP_DOMAIN}&count=100&extended=1"

func NewVkMock() http.RoundTripper {
	return &vkMock{}
}

type vkMock struct {
}

func (c *vkMock) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.String() != "https://api.vk.com/method/wall.get" {
		return nil, fmt.Errorf("VkMock: invalid url %s", req.URL.String())
	}

	req.ParseForm()
	domain := req.Form.Get("domain")
	filename := "data/" + domain + ".json"

	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	header := http.Header{}
	header.Set("Content-Type", "application/json")

	return &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        header,
		Body:          io.NopCloser(bytes.NewBuffer(buf)),
		ContentLength: int64(len(buf)),
		Request:       req,
	}, nil
}
