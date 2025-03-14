package paths

import (
	"context"
	"strings"

	"github.com/a-h/templ"
)

type (
	Name        string
	Params      map[string]string
	QueryParams map[string]string
)

func (n Name) String() string {
	return string(n)
}

type paths []Name

var Paths = []Name{
	APIHealth,
	APICollect,

	Home,
	About,

	CreateSubscription,
	UnSubscribe,
	VerifySubscriber,

	CreateAuthenticatedSession,
	StoreAuthenticatedSession,
	ForgotPassword,
	StoreForgotPassword,
	ResetPassword,
	StoreResetPassword,

	NewUser,
	CreateUser,
	VerifyEmail,

	Redirect,

	Dashboard,

	Robots,
	Sitemap,
	Favicon,
}

type Option func(*pathOptions)

type pathOptions struct {
	params      Params
	queryParams QueryParams
}

func WithParams(params Params) Option {
	return func(po *pathOptions) {
		po.params = params
	}
}

func WithQueryParams(queryParams QueryParams) Option {
	return func(po *pathOptions) {
		po.queryParams = queryParams
	}
}

func GP(
	ctx context.Context,
	name Name,
	opts ...Option,
) string {
	p, ok := ctx.Value(name).(string)
	if !ok {
		return ""
	}

	options := &pathOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if len(options.params) > 0 && strings.Contains(p, ":") {
		for key, value := range options.params {
			p = strings.Replace(p, ":"+key, value, 1)
		}
	}

	if len(options.queryParams) > 0 {
		var qp string
		for key, value := range options.queryParams {
			q := key + "=" + value
			if qp == "" {
				qp = q
			} else {
				qp = qp + "&" + q
			}
		}
		if qp != "" {
			p = p + "?" + qp
		}
	}

	return p
}

func GSP(
	ctx context.Context,
	name Name,
	opts ...Option,
) templ.SafeURL {
	return templ.SafeURL(GP(ctx, name, opts...))
}
