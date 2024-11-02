package utils

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

func ReadJsonFromFixture[T interface{}](filePath string) T {
	file, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = file.Close()
	}()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	var result T
	err = json.Unmarshal(byteValue, &result)
	if err != nil {
		panic(err)
	}

	return result
}
