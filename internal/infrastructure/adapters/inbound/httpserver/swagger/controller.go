package swagger

import (
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Controller struct{}

func NewController() *Controller {
	return &Controller{}
}

// RegisterRoutes registers routes for Controller
func (h *Controller) RegisterRoutes(router chi.Router) {
	router.Get("/swagger/*", httpSwagger.WrapHandler)
}