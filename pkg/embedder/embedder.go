// Package embedder defines the Embedder abstraction used by the rest of the
// service to turn free-form text into dense vectors. Concrete providers (e.g.
// Voyage) live in sibling files.
package embedder

import "context"

// InputType hints the provider whether the text is being embedded as a search
// query or as a document to be indexed. Most providers treat "" as neutral.
type InputType string

const (
	InputTypeNone     InputType = ""
	InputTypeQuery    InputType = "query"
	InputTypeDocument InputType = "document"
)

// Embedder converts a single text input into a vector.
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// TypedEmbedder is an optional extension for providers that accept an input
// type hint. Callers can type-assert to use it.
type TypedEmbedder interface {
	EmbedAs(ctx context.Context, text string, kind InputType) ([]float32, error)
}
