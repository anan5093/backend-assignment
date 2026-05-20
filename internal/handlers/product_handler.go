package handlers

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"backend-assignment/internal/services"
	"backend-assignment/internal/store"
	"backend-assignment/internal/utils"
	"backend-assignment/internal/validation"
)

type ProductHandler struct {
	service *services.ProductService
}

type createProductRequest struct {
	Name      string   `json:"name"`
	SKU       string   `json:"sku"`
	ImageURLs []string `json:"image_urls"`
	VideoURLs []string `json:"video_urls"`
}

type appendMediaRequest struct {
	ImageURLs *[]string `json:"image_urls"`
	VideoURLs *[]string `json:"video_urls"`
}

func NewProductHandler(service *services.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createProductRequest
	if err := utils.DecodeJSON(r.Body, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	product, err := h.service.Create(services.CreateProductInput{
		Name:      req.Name,
		SKU:       req.SKU,
		ImageURLs: req.ImageURLs,
		VideoURLs: req.VideoURLs,
	})
	if err != nil {
		if errors.Is(err, store.ErrDuplicateSKU) {
			utils.WriteError(w, http.StatusConflict, "duplicate sku")
			return
		}
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusCreated, product)
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := validation.ValidatePagination(r.URL.Query().Get("limit"), r.URL.Query().Get("offset"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	items, pagination := h.service.List(limit, offset)
	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"items":      items,
		"pagination": pagination,
	})
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	product, err := h.service.GetByID(mux.Vars(r)["id"])
	if err != nil {
		if errors.Is(err, store.ErrProductNotFound) {
			utils.WriteError(w, http.StatusNotFound, "product not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) AppendMedia(w http.ResponseWriter, r *http.Request) {
	var req appendMediaRequest
	if err := utils.DecodeJSON(r.Body, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	imageURLs := []string(nil)
	videoURLs := []string(nil)
	if req.ImageURLs != nil {
		imageURLs = *req.ImageURLs
	}
	if req.VideoURLs != nil {
		videoURLs = *req.VideoURLs
	}

	product, err := h.service.AppendMedia(mux.Vars(r)["id"], services.AppendMediaInput{
		ImageURLs: imageURLs,
		VideoURLs: videoURLs,
	})
	if err != nil {
		if errors.Is(err, store.ErrProductNotFound) {
			utils.WriteError(w, http.StatusNotFound, "product not found")
			return
		}
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, product)
}
