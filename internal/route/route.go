// Package route decouples HTTP route registration from the kernel: modules
// that expose endpoints implement Routable and are discovered by the
// httpserver module. The kernel itself never imports gin.
package route

import "github.com/gin-gonic/gin"

// Routable is implemented by modules that register HTTP routes. The httpserver
// module calls RegisterRoutes on every started module that implements it,
// mounting the routes under the shared /api/v1 group. Route-providing modules
// must therefore be registered before httpserver in kernel.New.
type Routable interface {
	RegisterRoutes(rg *gin.RouterGroup)
}
