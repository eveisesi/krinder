package roundtripper

import "net/http"

func UserAgent(userAgent string, original http.RoundTripper) http.RoundTripper {
	return roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		r.Header.Set("User-Agent", userAgent)
		if original == nil {
			original = http.DefaultTransport
		}

		return original.RoundTrip(r)
	})
}

type roundTripperFunc func(r *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
