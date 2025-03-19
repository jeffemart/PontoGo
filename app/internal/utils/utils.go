package utils

import (
	"encoding/base64"
	"fmt"
)

// DecodeBase64 decodifica uma string codificada em Base64
func DecodeBase64(encoded string) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("erro ao decodificar a string: %v", err)
	}
	return string(decodedBytes), nil
}
