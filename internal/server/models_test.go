package server

import (
	"bytes"
	"net/http/httptest"
	"testing"
)

func TestParseRequestAcceptsListingLinkAndPhotos(t *testing.T) {
	req := httptest.NewRequest("POST", "/listing-summary", bytes.NewBufferString(`{"listing":"nice flat","link":"https://example.com/listing","photos":["https://example.com/1.jpg","https://example.com/2.jpg"]}`))

	input, err := parseRequest(req)
	if err != nil {
		t.Fatalf("parseRequest() error = %v", err)
	}

	if input.Listing != "nice flat" {
		t.Fatalf("Listing = %q, want %q", input.Listing, "nice flat")
	}
	if input.Link != "https://example.com/listing" {
		t.Fatalf("Link = %q, want %q", input.Link, "https://example.com/listing")
	}
	if len(input.Photos) != 2 {
		t.Fatalf("len(Photos) = %d, want %d", len(input.Photos), 2)
	}
}

func TestParseRequestRejectsMissingLink(t *testing.T) {
	req := httptest.NewRequest("POST", "/listing-summary", bytes.NewBufferString(`{"listing":"nice flat"}`))

	if _, err := parseRequest(req); err == nil {
		t.Fatal("parseRequest() error = nil, want error")
	}
}
