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
