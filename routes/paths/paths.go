package paths

import (
	"context"
	"strings"

	"github.com/a-h/templ"
)

type RouteCtxKey string

type Path struct {
	Name string
	URL  string
}

var registry = make(map[string]Path)

func register(p Path) Path {
	registry[p.Name] = p
	return p
}

func GetAllPaths() []Path {
	paths := make([]Path, 0, len(registry))
	for _, p := range registry {
		paths = append(paths, p)
	}

	return paths
}

type (
	Params      map[string]string
	QueryParams map[string]string
)

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
	path Path,
	opts ...Option,
) string {
	p, ok := ctx.Value(RouteCtxKey(path.Name)).(string)
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
	path Path,
	opts ...Option,
) templ.SafeURL {
	return templ.SafeURL(GP(ctx, path, opts...))
}
