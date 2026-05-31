package localapi

import "net/http"

func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	jsonResponse(w, http.StatusOK, StatusResponse{OK: true, Version: defaultVersion})
}
