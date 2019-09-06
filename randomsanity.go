// appengine-based server to sanity check byte arrays
// that are supposed to be random.
package main

import (
	"encoding/hex"
	"fmt"
	"google.golang.org/appengine"
	"net/http"
	"strings"
	"time"
)

func debugHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")

	// Code useful for development/testing:
	fmt.Fprintf(w, "IPKey for memcache: %s\n",IPKey("q", r.RemoteAddr))

	fmt.Fprint(w, "***r.Header headers***\n")
	r.Header.Write(w)

	//	ctx := appengine.NewContext(r)
	//	fmt.Fprint(w, "Usage data:\n")
	//	for _, u := range GetUsage(ctx) {
	//		fmt.Fprintf(w, "%s,%d\n", u.Key, u.N)
	//	}
}

func submitBytesHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		http.Error(w, "Invalid GET", http.StatusBadRequest)
		return
	}
	b, err := hex.DecodeString(parts[len(parts)-1])
	if err != nil {
		http.Error(w, "Invalid hex", http.StatusBadRequest)
		return
	}
	// Need at least 16 bytes to hit the 1-in-2^60 false positive rate
	if len(b) < 16 {
		http.Error(w, "Must provide 16 or more bytes", http.StatusBadRequest)
		return
	}

	ctx := appengine.NewContext(r)


	// Rate-limit by IP address.
	var ratelimit uint64 = 60
	limited, err := RateLimitResponse(ctx, w, IPKey("q", r.RemoteAddr), ratelimit, time.Hour)
	if err != nil || limited {
		return
	}

	w.Header().Add("Content-Type", "application/json")

	// Returns some randomness caller can use to mix in to
	// their PRNG:
	addEntropyHeader(w)

	// First, some simple tests for non-random input:
	result, reason := LooksRandom(b)
	if !result {
		RecordUsage(ctx, "Fail_"+reason, 1)
		fmt.Fprint(w, "false")
		return
	}

	// Try to catch two machines with insufficient starting
	// entropy generating identical streams of random bytes.
	if len(b) > 64 {
		b = b[0:64] // Prevent DoS from excessive datastore lookups
	}
	unique, err := looksUnique(ctx, w, b)
	if err != nil {
		return
	}
	if unique {
		RecordUsage(ctx, "Success", 1)
		fmt.Fprint(w, "true")
	} else {
		RecordUsage(ctx, "Fail_Nonunique", 1)
		fmt.Fprint(w, "false")
	}
}
