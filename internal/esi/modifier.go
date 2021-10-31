package esi

import "net/http"

// Modifier funcs are trigger after data has been unmarshalled onto the interface
// but before the interface{} is cached in redis
type responseFunc func(out *Out)
type requestFunc func(req *http.Request)
