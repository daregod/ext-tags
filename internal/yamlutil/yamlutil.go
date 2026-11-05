// Package yamlutil contains small helpers for user-edited YAML files.
package yamlutil

import "bytes"

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// TrimBOM removes a UTF-8 byte-order mark. CRLF is handled by the YAML parser.
func TrimBOM(data []byte) []byte {
	return bytes.TrimPrefix(data, utf8BOM)
}
