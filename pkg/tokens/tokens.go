package tokens

import (
	"os"
)

var tokenSigningKey = []byte(os.Getenv("TOKEN_SIGNING_KEY"))
