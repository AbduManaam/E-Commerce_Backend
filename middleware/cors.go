
// CORSMiddleware configures CORS for the application
package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORSMiddleware configures CORS for the application
func CORSMiddleware() fiber.Handler {
	return cors.New(cors.Config{
		// Allow specific origins
		AllowOrigins: "http://localhost:5173,http://localhost:5175,http://localhost:3000",
		
		// Allow methods
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		
		// Allow headers
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Requested-With",
		
		// Expose headers
		ExposeHeaders: "Authorization",
		
		// Allow credentials (important for cookies)
		AllowCredentials: true,
		
		// Max age for preflight requests
		MaxAge: 86400, // 24 hours
	})
}