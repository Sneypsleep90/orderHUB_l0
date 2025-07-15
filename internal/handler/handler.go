package handler

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"orderHub_L0/internal/service"
	"path/filepath"
	"strings"
)

type Handler struct {
	svc    *service.Service
	logger *zap.Logger
}

func NewHandler(svc *service.Service, logger *zap.Logger) *Handler {
	return &Handler{
		svc:    svc,
		logger: logger,
	}
}

func (h *Handler) RegisterRoutes() *mux.Router {
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/order/{order_uid}", h.getOrder).Methods("GET")
	router.PathPrefix("/").HandlerFunc(h.serveWeb).Methods("GET")
	return router

}

func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	orderUID := mux.Vars(r)["order_uid"]
	order, err := h.svc.GetOrder(r.Context(), orderUID)
	if err != nil {
		h.logger.Error("failed to get order", zap.String("order_uid", orderUID), zap.Error(err))
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)

	}
}
func (h *Handler) serveWeb(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" || path == "" {
		path = "/index.html"
	}
	filePath := filepath.Join("web", strings.TrimPrefix(path, "/"))
	if _, err := filepath.Abs(filePath); err != nil {
		h.logger.Error("invalid file path", zap.String("path", filePath), zap.Error(err))
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	http.ServeFile(w, r, filePath)
}
