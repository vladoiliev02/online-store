package controller

import (
	"context"
	"log"
	"net/http"
)

const (
	UserIDKey = "userID"
)

type CtxKey string

func numericPathVariableExtractor(varName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			value, err := getNumericPathVariable(r, varName)
			if err != nil {
				log.Println(varName + " not found in request path")
				writeError(&HTTPError{
					Code:    http.StatusBadRequest,
					Message: "Invalid path parameter",
					Err:     err,
				}, w)
				return
			}
			ctx := context.WithValue(r.Context(), CtxKey(varName), value)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetContextParam[T any](varName string, ctx context.Context) T {
	val, ok := ctx.Value(CtxKey(varName)).(T)
	if !ok {
		var invalid T
		return invalid
	}
	return val
}

func SetContextParam(varName string, value any, ctx context.Context) context.Context {
	return context.WithValue(ctx, CtxKey(varName), value)
}
