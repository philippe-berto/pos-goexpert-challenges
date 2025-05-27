# Temperature System by ZIP Code with OTEL tracing

## Objective

Develop a system in Go that receives a ZIP code, identifies the city, and returns the current weather (temperature in Celsius, Fahrenheit, and Kelvin) along with the city name. This system must implement OTEL (OpenTelemetry) and Zipkin.

Based on the well-known scenario "Temperature System by ZIP Code" called **Service B**, a new project called **Service A** will be included.

---

## Testing

Just run

`docker-compose up -d`

and make a post request to

`http://localhost:8080 {"cep": "<zipcode>"}`

Access jaeger to see tracing in `http://localhost:16686/`

## Requirements

### Service A (Responsible for Input)

- The system must receive an 8-digit input via POST, using the schema:  
  `{ "cep": "29902555" }`
- The system must validate if the input is valid (contains 8 digits) and is a **STRING**
- If valid, it should be forwarded to Service B via HTTP
- If not valid, it must return:
  - **HTTP Code:** 422
  - **Message:** `invalid zipcode`

---

### Service B (Responsible for Orchestration)

- The system must receive a valid 8-digit ZIP code
- The system must search for the ZIP code and find the location name, then return the temperatures formatted in: Celsius, Fahrenheit, Kelvin, along with the location name.
- The system must respond appropriately in the following scenarios:

#### On success:

- **HTTP Code:** 200
- **Response Body:**  
  `{ "city": "São Paulo", "temp_C": 28.5, "temp_F": 83.3, "temp_K": 301.7 }`

#### On failure (ZIP code with correct format, but invalid):

- **HTTP Code:** 422
- **Message:** `invalid zipcode`

#### On failure (ZIP code not found):

- **HTTP Code:** 404
- **Message:** `can not find zipcode`

---

## Observability

After implementing the services, add OTEL + Zipkin:

- Implement distributed tracing between Service A and Service B
- Use spans to measure the response time of the ZIP code lookup and temperature lookup services

---

## Tips

- Use the [viaCEP API](https://viacep.com.br/) (or similar) to find the location.
- Use the [WeatherAPI](https://www.weatherapi.com/) (or similar) to get the temperatures.
- To convert Celsius to Fahrenheit:  
  `F = C * 1.8 + 32`
- To convert Celsius to Kelvin:  
  `K = C + 273`
- For OTEL implementation questions, see the [official documentation](https://opentelemetry.io/).
- For spans, see [this guide](https://opentelemetry.io/docs/instrumentation/go/manual/).
- You will need to use an OTEL collector service.
- For more information about Zipkin, see [here](https://zipkin.io/).

---

## Delivery

- The complete source code of the implementation.
- Documentation explaining how to run the project in a development environment.
- Use **docker/docker-compose** so we can test your application.
