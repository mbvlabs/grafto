package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/maypok86/otter"
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

type MW struct {
	rateLimiter otter.Cache[string, int32]
}

func New() MW {
	rateLimitCacheBuilder, err := otter.NewBuilder[string, int32](10_000)
	if err != nil {
		panic(err)
	}

	rateLimit, err := rateLimitCacheBuilder.WithTTL(1 * time.Minute).Build()
	if err != nil {
		panic(err)
	}

	return MW{
		rateLimit,
	}
}
