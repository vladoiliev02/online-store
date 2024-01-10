package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"online-store/dao"
	"strconv"

	"github.com/go-chi/chi/v5"
)

const (
	SessionKey = "sessionKey"
)

func Router() chi.Router {
	r := chi.NewRouter()

	r.Mount("/products", newProductRouter())
	r.Mount("/orders", newOrderRouter())
	r.Mount("/users", newUserRouter())

	r.Get("/liveness", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Get("/readiness", func(w http.ResponseWriter, r *http.Request) {
		isReady := dao.GetDAO().IsReady()
		log.Println("Checking readiness: ", isReady)
		if isReady {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	})

	return r
}

type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"error"`
}

type HTTPResponse[T any] struct {
	StatusCode int
	HasBody    bool
	Body       T
}

func NewOKResponse[T any](body T) *HTTPResponse[T] {
	return &HTTPResponse[T]{
		StatusCode: http.StatusOK,
		HasBody:    true,
		Body:       body,
	}
}

func NewResponse[T any](code int, body T) *HTTPResponse[T] {
	return &HTTPResponse[T]{
		StatusCode: code,
		HasBody:    true,
		Body:       body,
	}
}

func NewStatusResponse[T any](code int) *HTTPResponse[T] {
	return &HTTPResponse[T]{
		StatusCode: code,
		HasBody:    false,
	}
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s: %v", e.Code, e.Message, e.Err)
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

func ControllerHandler[T any](handler func(*http.Request) (*HTTPResponse[T], error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response, err := handler(r)
		if err != nil {
			writeError(err, w)
		} else {
			writeResponse(response, w)
		}
	}
}

func writeResponse[T any](response *HTTPResponse[T], w http.ResponseWriter) {
	if response.HasBody {
		log.Println("writing response")
		responseJSON, err := json.Marshal(response.Body)
		if err != nil {
			internalError(w)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(responseJSON)
	}

	w.WriteHeader(response.StatusCode)
}

func writeError(err error, w http.ResponseWriter) {
	e, ok := err.(*HTTPError)
	if !ok {
		e = &HTTPError{Code: http.StatusInternalServerError, Message: "Internal Server Error", Err: err}
	}

	log.Println(err.Error() + " - caused by - " + e.Err.Error())
	httpErr, err := json.Marshal(e)
	if err != nil {
		internalError(w)
		return
	}
	log.Println("Returning:", string(httpErr))

	http.Error(w, string(httpErr), e.Code)
	w.Header().Set("Content-Type", "application/json")
}

func internalError(w http.ResponseWriter) {
	http.Error(w, `{"message":"internal-server-error"}`, http.StatusInternalServerError)
}

func toInt(str string) (int64, error) {
	i, err := strconv.ParseInt(str, 10, 64)
	return i, err
}

func getQueryParam(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

func getNumericQueryParam(r *http.Request, name string) (int64, error) {
	val := r.URL.Query().Get(name)
	if val == "" {
		return 0, nil
	} else {
		return toInt(val)
	}
}

func getNumericPathVariable(r *http.Request, name string) (int64, error) {
	str := chi.URLParam(r, name)
	return toInt(str)
}

func jsonUnmarshalBody[T any](r *http.Request) (*T, error) {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, &HTTPError{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
			Err:     err,
		}
	}

	var obj T
	err = json.Unmarshal(bytes, &obj)
	if err != nil {
		return nil, &HTTPError{
			Code:    http.StatusBadRequest,
			Message: "Invalid json in request body",
			Err:     err,
		}
	}

	return &obj, nil
}
