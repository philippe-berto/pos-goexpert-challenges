//go:build tool
// +build tool

package server

//go:generate mockgen -source=quotes/usecase/usecase.go -destination=./mocks/usecase_mock.go -package=mocks GetterDollarQuote
