package factory

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_createDirector(t *testing.T) {
	testCases := []struct {
		name        string
		target      *url.URL
		req         *http.Request
		expectedReq *http.Request
	}{
		{
			name:   "base case",
			target: createURL(t, "https://portainer.io/api/docker?a=5&b=6"),
			req: createRequest(
				t,
				"GET",
				"https://agent-portainer.io/test?c=7",
				map[string]string{"Accept-Encoding": "gzip", "Accept": "application/json", "User-Agent": "something"},
			),
			expectedReq: createRequest(
				t,
				"GET",
				"https://portainer.io/api/docker/test?a=5&b=6&c=7",
				map[string]string{"Accept-Encoding": "gzip", "Accept": "application/json", "User-Agent": "something"},
			),
		},
		{
			name:   "no User-Agent",
			target: createURL(t, "https://portainer.io/api/docker?a=5&b=6"),
			req: createRequest(
				t,
				"GET",
				"https://agent-portainer.io/test?c=7",
				map[string]string{"Accept-Encoding": "gzip", "Accept": "application/json"},
			),
			expectedReq: createRequest(
				t,
				"GET",
				"https://portainer.io/api/docker/test?a=5&b=6&c=7",
				map[string]string{"Accept-Encoding": "gzip", "Accept": "application/json", "User-Agent": ""},
			),
		},
		{
			name:   "Sensitive Headers",
			target: createURL(t, "https://portainer.io/api/docker?a=5&b=6"),
			req: createRequest(
				t,
				"GET",
				"https://agent-portainer.io/test?c=7",
				map[string]string{
					"Accept-Encoding": "gzip",
					"Accept":          "application/json",
					"User-Agent":      "something",
					"Cookie":          "junk",
					"X-Csrf-Token":    "junk",
				},
			),
			expectedReq: createRequest(
				t,
				"GET",
				"https://portainer.io/api/docker/test?a=5&b=6&c=7",
				map[string]string{"Accept-Encoding": "gzip", "Accept": "application/json", "User-Agent": "something"},
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			director := createDirector(tc.target)
			director(tc.req)

			if diff := cmp.Diff(tc.req, tc.expectedReq, cmp.Comparer(compareRequests)); diff != "" {
				t.Fatalf("requests are different: \n%s", diff)
			}
		})
	}
}

func createURL(t *testing.T, urlString string) *url.URL {
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		t.Fatalf("Failed to create url: %s", err)
	}

	return parsedURL
}

func createRequest(t *testing.T, method, url string, headers map[string]string) *http.Request {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("Failed to create http request: %s", err)
	} else {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}

	return req
}

func compareRequests(a, b *http.Request) bool {
	methodEqual := a.Method == b.Method
	urlEqual := cmp.Diff(a.URL, b.URL) == ""
	hostEqual := a.Host == b.Host
	protoEqual := a.Proto == b.Proto && a.ProtoMajor == b.ProtoMajor && a.ProtoMinor == b.ProtoMinor
	headersEqual := cmp.Diff(a.Header, b.Header) == ""

	return methodEqual && urlEqual && hostEqual && protoEqual && headersEqual
}
