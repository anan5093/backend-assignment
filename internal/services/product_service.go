package services

import (
	"errors"
	"strings"
	"time"

	"backend-assignment/internal/models"
	"backend-assignment/internal/store"
	"backend-assignment/internal/validation"
)

type ProductService struct {
	store *store.ProductStore
}

type CreateProductInput struct {
	Name      string
	SKU       string
	ImageURLs []string
	VideoURLs []string
}

type AppendMediaInput struct {
	ImageURLs []string
	VideoURLs []string
}

func NewProductService(store *store.ProductStore) *ProductService {
	return &ProductService{store: store}
}

func (s *ProductService) Create(input CreateProductInput) (models.Product, error) {
	name := strings.TrimSpace(input.Name)
	sku := strings.TrimSpace(input.SKU)

	if err := validation.RequiredString(name, "name"); err != nil {
		return models.Product{}, err
	}
	if err := validation.RequiredString(sku, "sku"); err != nil {
		return models.Product{}, err
	}
	if err := validation.ValidateURLs(input.ImageURLs, "image_urls"); err != nil {
		return models.Product{}, err
	}
	if err := validation.ValidateURLs(input.VideoURLs, "video_urls"); err != nil {
		return models.Product{}, err
	}

	return s.store.Create(name, sku, input.ImageURLs, input.VideoURLs, time.Now().UTC())
}

func (s *ProductService) List(limit, offset int) ([]models.ProductSummary, models.Pagination) {
	items, total := s.store.List(limit, offset)
	return items, models.Pagination{
		Limit:  limit,
		Offset: offset,
		Total:  total,
	}
}

func (s *ProductService) GetByID(id string) (models.Product, error) {
	return s.store.GetByID(id)
}

func (s *ProductService) AppendMedia(id string, input AppendMediaInput) (models.Product, error) {
	if len(input.ImageURLs) == 0 && len(input.VideoURLs) == 0 {
		return models.Product{}, errors.New("at least one media URL is required")
	}
	if err := validation.ValidateURLs(input.ImageURLs, "image_urls"); err != nil {
		return models.Product{}, err
	}
	if err := validation.ValidateURLs(input.VideoURLs, "video_urls"); err != nil {
		return models.Product{}, err
	}

	return s.store.AppendMedia(id, input.ImageURLs, input.VideoURLs, validation.MaxURLCount)
}
