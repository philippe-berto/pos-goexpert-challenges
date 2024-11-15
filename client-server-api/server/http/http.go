package httpserver

import (
	"encoding/json"
	"log"
	"net/http"
	"server/quotes/usecase"

	"github.com/go-chi/chi/v5"
)

type (
	handler struct {
		usecase usecase.GetterDollarQuote
		Router  chi.Router
	}

	response struct {
		Err   *string
		Value *string
	}
)

func New(u usecase.GetterDollarQuote) *handler {
	r := chi.NewRouter()
	handler := &handler{
		usecase: u,
		Router:  r,
	}
	r.Get("/cotacao", handler.getDollarQuote)
	return handler
}

func (h *handler) getDollarQuote(w http.ResponseWriter, r *http.Request) {
	response := response{Err: nil}
	quote, err := h.usecase.GetDollarQuote()
	if err != nil {
		log.Println("Error getting quote:", err)
		errValue := err.Error()
		response.Err = &errValue
	}
	response.Value = quote
	json.NewEncoder(w).Encode(response)
}
