## How to run

```
docker build -t cep-service -f build/Dockerfile .
docker run --rm -p 8080:8080 cep-service
```

Then call

```
http://localhost:8080/{zip-code}
```

The Cloud Run Version is

```
https://pos-goexpert-challenges-818603360016.europe-west1.run.app/{zip-code}
```

## Objective

Develop a Go system that receives a Brazilian ZIP code (CEP), identifies the city, and returns the current weather (temperature in Celsius, Fahrenheit, and Kelvin). This system must be deployed on Google Cloud Run.

## Requirements

- The system must receive a valid 8-digit ZIP code (CEP).
- The system must look up the ZIP code, find the location name, and return the temperatures formatted as: Celsius, Fahrenheit, Kelvin.
- The system must respond appropriately in the following scenarios:
  - **Success:**
    - HTTP Code: 200
    - Response Body: `{ "temp_C": 28.5, "temp_F": 28.5, "temp_K": 28.5 }`
  - **Failure, if the ZIP code is invalid (wrong format):**
    - HTTP Code: 422
    - Message: `invalid zipcode`
  - **Failure, if the ZIP code is not found:**
    - HTTP Code: 404
    - Message: `can not find zipcode`
- The system must be deployed on Google Cloud Run.

## Tips

- Use the viaCEP API (or similar) to find the location for the ZIP code: https://viacep.com.br/
- Use the WeatherAPI (or similar) to get the desired temperatures: https://www.weatherapi.com/
- To convert Celsius to Fahrenheit, use the formula: `F = C * 1.8 + 32`
- To convert Celsius to Kelvin, use the formula: `K = C + 273`
  - Where F = Fahrenheit
  - Where C = Celsius
  - Where K = Kelvin

## Deliverables

- Complete source code of the implementation.
- Automated tests demonstrating functionality.
- Use docker/docker-compose so your application can be tested.
- Deployment on Google Cloud Run (free tier) with an active endpoint for access.
