package auth

import (
	"context"
	"dnd_back/api"
	"fmt"
	"log"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3filter"
	middleware "github.com/oapi-codegen/nethttp-middleware"
)

func CreateMiddleware(v JWSValidator) (func(next http.Handler) http.Handler, error) {
	spec, err := api.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("loading spec: %w", err)
	}

	spec.Servers = nil

	validator := middleware.OapiRequestValidatorWithOptions(spec,
		&middleware.Options{
			SilenceServersWarning: true,
			ErrorHandlerWithOpts: func(ctx context.Context, err error, w http.ResponseWriter, r *http.Request, opts middleware.ErrorHandlerOpts) {
				log.Printf("request validation error: method=%s path=%s status=%d message=%s", r.Method, r.URL.String(), opts.StatusCode, err.Error())
				http.Error(w, err.Error(), opts.StatusCode)
			},
			Options: openapi3filter.Options{
				AuthenticationFunc: NewAuthenticator(v),
			},
		})

	return validator, nil
}
