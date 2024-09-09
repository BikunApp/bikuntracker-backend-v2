package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func ParseRequestBody[T interface{}](body io.ReadCloser) (res T, err error) {
	bodyInByte, err := io.ReadAll(body)
	if err != nil {
		err = fmt.Errorf("unable to read request body: %w", err)
		return
	}

	err = json.Unmarshal(bodyInByte, &res)
	if err != nil {
		err = fmt.Errorf("Unable to parse request body: %w", err)
		return
	}

	return
}

func ParseResponseBody[T interface{}](response *http.Response) (body T, err error) {
	if !(response.StatusCode >= 200 && response.StatusCode < 300) {
		err = fmt.Errorf("response returned a non 2xx status code")
		return
	}

	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("unable to read response body: %w", err)
		return
	}

	err = json.Unmarshal(respBody, &body)
	if err != nil {
		err = fmt.Errorf("unable to unmarshal response body: %w", err)
		return
	}

	return
}
