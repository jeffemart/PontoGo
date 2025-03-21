package utils

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
)

// Logger é o logger global
var Logger *log.Logger

// InitLogger inicializa o logger
func InitLogger() {
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	Logger = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)
}

// DecodeBase64 decodifica uma string codificada em Base64
func DecodeBase64(encoded string) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		Logger.Printf("Erro ao decodificar a string: %v", err)
		return "", fmt.Errorf("erro ao decodificar a string: %v", err)
	}
	return string(decodedBytes), nil
}
