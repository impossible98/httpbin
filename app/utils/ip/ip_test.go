package ip

import (
	// import built-in packages
	"net/http"
	"testing"
)

func Test_GetClientIP(t *testing.T) {
	t.Parallel()

	makeHeaders := func(m map[string]string) http.Header {
		h := make(http.Header, len(m))
		for k, v := range m {
			h.Set(k, v)
		}
		return h
	}

	testCases := map[string]struct {
		given *http.Request
		want  string
	}{
		"custom platform headers take precedence": {
			given: &http.Request{
				Header: makeHeaders(map[string]string{
					"Fly-Client-IP":   "9.9.9.9",
					"X-Forwarded-For": "1.1.1.1,2.2.2.2,3.3.3.3",
				}),
				RemoteAddr: "0.0.0.0",
			},
			want: "9.9.9.9",
		},
		"x-forwarded-for is parsed": {
			given: &http.Request{
				Header: makeHeaders(map[string]string{
					"X-Forwarded-For": "1.1.1.1,2.2.2.2,3.3.3.3",
				}),
				RemoteAddr: "0.0.0.0",
			},
			want: "1.1.1.1",
		},
		"remoteaddr is fallback": {
			given: &http.Request{
				RemoteAddr: "0.0.0.0",
			},
			want: "0.0.0.0",
		},
	}
	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := GetClientIP(tc.given); got != tc.want {
				t.Errorf("getClientIP() = %v, want %v", got, tc.want)
			}
		})
	}
}
