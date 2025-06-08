package helpers

import "net/http"

type LangInfo struct {
	Size        int
	Occurrences int
	Colour      string
	Prop        float64
}

type TransportWithToken struct {
	Token     string
	Transport http.RoundTripper
}
