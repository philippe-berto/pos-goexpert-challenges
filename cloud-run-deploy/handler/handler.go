package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/philippe-berto/pos-goexpert-challenges/cloud-run-deploy/service"
)

type (
	router struct {
		routes map[string]map[string]http.HandlerFunc
	}
	Handler struct {
		s service.Cep
		r *router
	}
)

func New(ctx context.Context) (*Handler, error) {
	service, err := service.New(ctx, true)
	if err != nil {
		panic(err)
	}

	router := &router{
		routes: make(map[string]map[string]http.HandlerFunc),
	}

	return &Handler{
		s: *service,
		r: router,
	}, nil
}

func (h *Handler) GetWeather(w http.ResponseWriter, req *http.Request) {
	cep := Param(req, "cep")
	if cep == "" {
		http.Error(w, "CEP is required", http.StatusBadRequest)
		return
	}

	result, err := h.s.GetWeather(cep)
	if err != nil {
		switch err.Error() {
		case "WRONG_FORMAT":
			http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)

			return
		case "NOT_FOUND":
			http.Error(w, "can not find zipcode", http.StatusNotFound)

			return
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func Param(r *http.Request, key string) string {
	params, ok := r.Context().Value("params").(map[string]string)
	if !ok {
		return ""
	}
	return params[key]
}
