package contexts

import (
	"context"
)

func ExtractApp(ctx context.Context) App {
	appCtx, ok := ctx.Value(AppKey{}).(App)
	if !ok {
		return App{}
	}

	return appCtx
}

func ExtractFlashMessages(ctx context.Context) []FlashMessage {
	value, ok := ctx.Value(FlashKey{}).([]FlashMessage)
	if !ok {
		return nil
	}

	return value
}
