package middleware

import (
	"net/http"
)

func ForwardedForMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// clientIP := r.RemoteAddr
		// if strings.Contains(clientIP, "[") || strings.Contains(clientIP, "]") {
		// 	// Get all between []
		// 	clientIP = strings.Split(clientIP, "[")[1]
		// 	clientIP = strings.Split(clientIP, "]")[0]
		// } else {
		// 	// Get all before :
		// 	clientIP = strings.Split(clientIP, ":")[0]
		// }

		// // Get the existing X-Forwarded-For header
		// existingForwardedFor := r.Header.Get("X-Forwarded-For")
		// if existingForwardedFor != "" {
		// 	// Append the new client IP to the existing header
		// 	r.Header.Set("X-Forwarded-For", existingForwardedFor+", "+clientIP)
		// } else {
		// 	// Set the new client IP as the header
		// 	r.Header.Set("X-Forwarded-For", clientIP)
		// }

		next.ServeHTTP(w, r)
	})
}
