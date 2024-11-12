package httpserver

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"server/mocks"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type mockHolder struct {
	http.Handler
	mockS *mocks.MockGetterDollarQuote
}

func TestGetDollarQuote(t *testing.T) {
	t.Run("should return dollar quote", func(t *testing.T) {
		mockHolder, responseWriter := setupTest(t)
		quote := "5.30"
		mockHolder.mockS.EXPECT().GetDollarQuote().Return(&quote, nil)

		req, err := http.NewRequest("GET", "/cotacao", nil)
		if err != nil {
			t.Fatal(err)
		}

		mockHolder.ServeHTTP(responseWriter, req)
		log.Printf("Response: %v", responseWriter.Body)

		var actual response
		err = json.NewDecoder(responseWriter.Body).Decode(&actual)
		if err != nil {
			t.Fatal(err)
		}

		expected := response{Err: nil, Value: &quote}

		assert.Equal(t, http.StatusOK, responseWriter.Code)
		assert.Equal(t, expected, actual)
	})
}

func setupTest(t *testing.T) (*mockHolder, *httptest.ResponseRecorder) {
	ctrl := gomock.NewController(t)
	responseWriter := httptest.NewRecorder()

	mockS := mocks.NewMockGetterDollarQuote(ctrl)

	h := New(mockS)
	r := chi.NewRouter()
	r.Mount("/", h.Router)

	mockHolder := &mockHolder{
		Handler: r,
		mockS:   mockS,
	}

	return mockHolder, responseWriter
}
