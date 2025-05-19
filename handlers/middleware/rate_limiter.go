package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/views/authentication"
)

func (m MW) LoginRateLimiter() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			ip := c.RealIP()

			hits, found := m.rateLimiter.Get(ip)
			if !found {
				if ok := m.rateLimiter.Set(ip, 1); !ok {
					return next(c)
				}
			}
			if hits <= 5 {
				if ok := m.rateLimiter.Set(ip, hits+1); !ok {
					return next(c)
				}
			}

			if hits > 5 {
				c.Response().
					Header().
					Set("HX-Retarget", "div[id='login-flag']")
				c.Response().
					Header().
					Set("HX-Reswap", "outerHTML")

				return authentication.LoginError("Too many failed attemps!").
					Render(c.Request().Context(), c.Response())
			}

			return next(c)
		}
	}
}
