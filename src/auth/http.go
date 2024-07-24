package auth

import (
	"net/http"

	apitypes "github.com/secnex/secnex-api-gateway/types"
)

func CheckAuthentication(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("Authorization") == "" {
		result := apitypes.ResultError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
			Error:   "Authorization header not present",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(result.String()))
		return false
	}

	return true
}
