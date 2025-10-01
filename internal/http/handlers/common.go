package handlers

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

func requiredPostFormField(r *http.Request, key string) (string, error) {
	err := r.ParseForm()
	if err != nil {
		return "", errors.WithStack(err)
	}

	if !r.PostForm.Has(key) {
		return "", fmt.Errorf("required field: %s", key)
	}
	return r.PostForm.Get(key), nil
}

func requiredQueryField(r *http.Request, key string) (string, error) {
	if !r.URL.Query().Has(key) {
		return "", fmt.Errorf("required form field: %s", key)
	}
	return r.URL.Query().Get(key), nil
}

func requiredPathParam(r *http.Request, key string) (string, error) {
	value := r.PathValue(key)
	if value == "" {
		return "", fmt.Errorf("required path parameter: %s", key)
	}
	return value, nil
}
