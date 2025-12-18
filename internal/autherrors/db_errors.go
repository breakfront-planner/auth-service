package autherrors

import (
	"fmt"
)

func ErrMissingEnvVars(varNames []string) error {
	return fmt.Errorf("missing required environment variables: %v", varNames)
}
