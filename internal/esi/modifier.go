package esi

import "net/http"

// Modifier funcs are trigger after data has been unmarshalled onto the interface
// but before the interface{} is cached in redis
type responseFunc func(out *Out)
type RequestFunc func(req *http.Request)

func AddIfNoneMatchHeader(etag string) RequestFunc {
	return func(req *http.Request) {
		if etag == "" {
			return
		}
		req.Header.Set("If-None-Match", etag)
	}

}
