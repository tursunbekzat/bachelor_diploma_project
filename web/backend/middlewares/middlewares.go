package middlewares

import (
	"backend/data"
    
	"log"
	"net/http"
)

// JWTAuthMiddleware checks for a valid JWT token
func JWTAuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("token")
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        claims, err := data.ValidateJWT(cookie.Value)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        log.Printf("Authenticated user: %s", claims.Username)

        // Pass the request to the next handler
        next.ServeHTTP(w, r)
    })
}

