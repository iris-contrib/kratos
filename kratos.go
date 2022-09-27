package kratos

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	thttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/kataras/iris/v12"
)

const (
	baseContentType = "application"
)

// ContentType returns the content-type with base prefix.
func ContentType(subtype string) string {
	return strings.Join([]string{baseContentType, subtype}, "/")
}

// Error encodes the object to the HTTP response.
func Error(ctx iris.Context, err error) {
	if err == nil {
		return
	}

	codec, _ := thttp.CodecForRequest(ctx.Request(), "Accept")
	se := errors.FromError(err)
	body, err := codec.Marshal(se)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	code := int(se.Code)
	contentType := codec.Name()
	ctx.StatusCode(code)
	ctx.ContentType(contentType)
	ctx.Write(body)
}

// Middlewares return middlewares wrapper.
func Middlewares(m ...middleware.Middleware) iris.Handler {
	chain := middleware.Chain(m...)
	return func(ctx iris.Context) {
		next := func(stdCtx context.Context, req interface{}) (interface{}, error) {
			ctx.Next()
			var err error
			if ctx.ResponseWriter().StatusCode() >= iris.StatusBadRequest {
				err = errors.Errorf(ctx.ResponseWriter().StatusCode(), errors.UnknownReason, errors.UnknownReason)
			}
			return ctx.ResponseWriter(), err
		}

		next = chain(next)
		thttp.SetOperation(ctx, ctx.Path())
		next(ctx, ctx.Request())
	}
}
