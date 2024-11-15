## Multithreading Challenge

### Key words

go, golang, goroutine

### Running

- Just run `go run main.go`
- You can change the constant values `brasilApiTimeout` and `viaCepTimeout` to force API timeouts

### Description

"In this challenge, you will need to use what we have learned about Multithreading and APIs to get the fastest result between two different APIs.

The two requests will be made simultaneously to the following APIs:

https://brasilapi.com.br/api/cep/v1/01153000 + cep

http://viacep.com.br/ws/" + cep + "/json/

The requirements for this challenge are:

Accept the API that delivers the fastest response and discard the slower response.

The result of the request should be displayed on the command line with the address data, as well as which API sent it.

Limit the response time to 1 second. Otherwise, a timeout error should be displayed."
