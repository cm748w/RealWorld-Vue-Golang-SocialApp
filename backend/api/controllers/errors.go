package controllers

import "log"

// internalDetail logs the underlying error server-side and returns the generic
// text that goes into a client-facing response body.
//
// Why this exists: error responses used to embed the raw underlying error text,
// which leaked MongoDB collection names, driver messages and parser internals to
// any caller who could trigger a failure path. The real error is still recorded
// in the server log so operators keep their diagnostic trail.
func internalDetail(err error) string {
	if err != nil {
		log.Printf("request failed: %v", err)
	}
	return "internal server error"
}
