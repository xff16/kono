package kairyu

import "net/http"

type Context struct {
	Request  *http.Request
	Response *http.Response
	Route    *Route
	Data     map[string]interface{}
}
