package artifact

import (
	"encoding/json"
	"io"
)

// writeJSON writes data as JSON to the provided writer
// Uses Go's standard encoding/json package following the "standard library first" principle
func writeJSON(w io.Writer, data interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ") // Pretty print for readability
	return encoder.Encode(data)
}

// readJSON reads JSON data from the provided reader into the destination
// Uses Go's standard encoding/json package following the "standard library first" principle
func readJSON(r io.Reader, dest interface{}) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(dest)
}
