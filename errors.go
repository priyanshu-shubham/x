package main

import "errors"

// Configuration errors
var (
	ErrMissingAPIKey    = errors.New("API key is required")
	ErrMissingProjectID = errors.New("project ID is required for Vertex AI")
	ErrMissingRegion    = errors.New("region is required for Vertex AI")
	ErrInvalidAuthType  = errors.New("invalid authentication type")
	ErrInvalidChoice    = errors.New("invalid choice")
	ErrNoEditorFound    = errors.New("no editor found; set $EDITOR environment variable")
)
