package contexts

import "context"

func ExtractApp(ctx context.Context) App {
	appCtx, ok := ctx.Value(AppKey{}).(App)
	if !ok {
		return App{}
	}

	return appCtx
}
