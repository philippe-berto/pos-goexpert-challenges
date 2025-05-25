package internal

import "fmt"

func VerifyCep(cep string) error {
	if len(cep) != 8 {
		return fmt.Errorf("invalid CEP: %s", cep)
	}

	for _, char := range cep {
		if char < '0' || char > '9' {
			return fmt.Errorf("invalid CEP: %s", cep)
		}
	}

	return nil
}
