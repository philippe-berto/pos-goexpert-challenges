package service

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyCep(t *testing.T) {
	tests := []struct {
		cep      string
		expected bool
	}{
		{"22461000", true},
		{"12345678", true},
		{"1234567", false},
		{"123456789", false},
		{"1234abcd", false},
	}

	ctx := context.Background()
	cep := Cep{
		ctx: ctx,
	}

	for _, test := range tests {

		err := cep.verifyCep(test.cep)
		if (err == nil) != test.expected {
			t.Errorf("VerifyCep(%s) = %v; expected %v", test.cep, err == nil, test.expected)
		}
	}
}

func TestGetFromBrasilCep(t *testing.T) {
	ctx := context.Background()

	cep := Cep{
		ctx: ctx,
	}

	testCep := "22461000"
	expectedCity := "Rio de Janeiro"
	expectedState := "RJ"
	expectedNeighborhood := "Jardim Botânico"
	expectedStreet := "Rua Jardim Botânico"
	expectedCep := "22461000"

	result, err := cep.GetFromBrasilCep(testCep)
	if err != nil {
		t.Errorf("GetFromBrasilCep(%s) returned an error: %v", testCep, err)
	}

	assert.Equal(t, expectedCity, result.City)
	assert.Equal(t, expectedState, result.State)
	assert.Equal(t, expectedNeighborhood, result.Neighborhood)
	assert.Equal(t, expectedStreet, result.Street)
	assert.Equal(t, expectedCep, result.Cep)
}

func TestGetFromBrasilCepInvalid(t *testing.T) {
	ctx := context.Background()

	cep := Cep{
		ctx: ctx,
	}

	testCep := "12345678"

	_, err := cep.GetFromBrasilCep(testCep)
	assert.Equal(t, err.Error(), "invalid CEP: 12345678")
}

func TestGetTemperature(t *testing.T) {
	ctx := context.Background()

	cep := Cep{
		ctx: ctx,
	}

	testCity := "Rio de Janeiro"

	res, err := cep.GetTemperature(testCity)

	assert.NoError(t, err)
	assert.NotEqual(t, reflect.TypeOf(res.Current.TempC).Kind(), reflect.TypeOf(float64(0)))
}

func TestGetTemperatureInvalid(t *testing.T) {
	ctx := context.Background()

	cep := Cep{
		ctx: ctx,
	}

	testCity := "asdasdas"

	res, err := cep.GetTemperature(testCity)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get weather data")
	assert.Equal(t, res.Current.TempC, 0.0)
}
