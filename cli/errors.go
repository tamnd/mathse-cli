package cli

import (
	"errors"

	"github.com/tamnd/mathse-cli/mathse"
)

func isNotFound(err error) bool {
	return errors.Is(err, mathse.ErrNotFound)
}
