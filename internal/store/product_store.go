package store

import (
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"backend-assignment/internal/models"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrDuplicateSKU    = errors.New("duplicate sku")
)

type productRecord struct {
	ID        string
	Name      string
	SKU       string
	CreatedAt time.Time
}

type productMedia struct {
	ImageURLs []string
	VideoURLs []string
}

type ProductStore struct {
	mu               sync.RWMutex
	nextID           atomic.Uint64
	productsByID     map[string]*productRecord
	mediaByProductID map[string]*productMedia
	skuIndex         map[string]string
	productOrder     []string
}

func NewProductStore() *ProductStore {
	return &ProductStore{
		productsByID:     make(map[string]*productRecord),
		mediaByProductID: make(map[string]*productMedia),
		skuIndex:         make(map[string]string),
		productOrder:     make([]string, 0),
	}
}

func (s *ProductStore) Create(name, sku string, imageURLs, videoURLs []string, createdAt time.Time) (models.Product, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.skuIndex[sku]; exists {
		return models.Product{}, ErrDuplicateSKU
	}

	id := s.generateID()
	record := &productRecord{
		ID:        id,
		Name:      name,
		SKU:       sku,
		CreatedAt: createdAt,
	}

	s.productsByID[id] = record
	s.mediaByProductID[id] = &productMedia{
		ImageURLs: append([]string(nil), imageURLs...),
		VideoURLs: append([]string(nil), videoURLs...),
	}
	s.skuIndex[sku] = id
	s.productOrder = append(s.productOrder, id)

	return s.toProduct(record, s.mediaByProductID[id]), nil
}

func (s *ProductStore) List(limit, offset int) ([]models.ProductSummary, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.productOrder)
	if offset >= total {
		return []models.ProductSummary{}, total
	}

	end := offset + limit
	if end > total {
		end = total
	}

	items := make([]models.ProductSummary, 0, end-offset)
	for _, id := range s.productOrder[offset:end] {
		record := s.productsByID[id]
		media := s.mediaByProductID[id]

		thumbnailURL := ""
		if len(media.ImageURLs) > 0 {
			thumbnailURL = media.ImageURLs[0]
		}

		items = append(items, models.ProductSummary{
			ID:           record.ID,
			Name:         record.Name,
			SKU:          record.SKU,
			ImageCount:   len(media.ImageURLs),
			VideoCount:   len(media.VideoURLs),
			ThumbnailURL: thumbnailURL,
			CreatedAt:    record.CreatedAt,
		})
	}

	return items, total
}

func (s *ProductStore) GetByID(id string) (models.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.productsByID[id]
	if !ok {
		return models.Product{}, ErrProductNotFound
	}

	return s.toProduct(record, s.mediaByProductID[id]), nil
}

func (s *ProductStore) AppendMedia(id string, imageURLs, videoURLs []string, maxURLsPerType int) (models.Product, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.productsByID[id]
	if !ok {
		return models.Product{}, ErrProductNotFound
	}

	media := s.mediaByProductID[id]
	if len(media.ImageURLs)+len(imageURLs) > maxURLsPerType {
		return models.Product{}, errors.New("image_urls cannot contain more than 20 URLs")
	}
	if len(media.VideoURLs)+len(videoURLs) > maxURLsPerType {
		return models.Product{}, errors.New("video_urls cannot contain more than 20 URLs")
	}

	media.ImageURLs = append(media.ImageURLs, imageURLs...)
	media.VideoURLs = append(media.VideoURLs, videoURLs...)

	return s.toProduct(record, media), nil
}

func (s *ProductStore) generateID() string {
	return strconv.FormatUint(s.nextID.Add(1), 10)
}

func (s *ProductStore) toProduct(record *productRecord, media *productMedia) models.Product {
	return models.Product{
		ID:        record.ID,
		Name:      record.Name,
		SKU:       record.SKU,
		ImageURLs: append([]string(nil), media.ImageURLs...),
		VideoURLs: append([]string(nil), media.VideoURLs...),
		CreatedAt: record.CreatedAt,
	}
}
