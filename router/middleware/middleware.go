package middleware

import (
	"fmt"
	"strings"

	"github.com/mbvlabs/grafto/config"
)

var AuthenticatedSessionName = fmt.Sprintf(
	"ua-%s-%s",
	strings.ToLower(config.Cfg.ProjectName),
	config.Cfg.Environment,
)

const (
	FlashSessionKey     = "flash_messages"
	SessIsAuthenticated = "is_authenticated"
	SessUserID          = "user_id"
	SessUserEmail       = "user_email"
	SessIsAdmin         = "is_admin"
	oneWeekInSeconds    = 604800
)
