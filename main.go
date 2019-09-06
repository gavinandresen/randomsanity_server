package main

import (
	"google.golang.org/appengine"
	"net/http"
)

func main() {
	// Main API point, sanity check hex bytes
	http.HandleFunc("/v1/q/", submitBytesHandler)

	// Start an email loop to get an id token, to be
	// notified via email of failures:
	http.HandleFunc("/v1/registeremail/", registerEmailHandler)

	// Remove an id token
	http.HandleFunc("/v1/unregister/", unRegisterIDHandler)

	// Get usage stats
	http.HandleFunc("/v1/usage", usageHandler)

	// Development/testing...
	http.HandleFunc("/v1/debug", debugHandler)

	// Redirect to www. home page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "https://www.randomsanity.org/", 301)
	})

	appengine.Main()
}
