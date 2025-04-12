package repository

import (
	"context"

	"todolist/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryRepositoryAdapter struct {
	db *gorm.DB
}

func NewCategoryRepositoryAdapter(srcDB *gorm.DB) *CategoryRepositoryAdapter {
	return &CategoryRepositoryAdapter{
		db: srcDB,
	}
}

type Category struct {
	ID     uuid.UUID `gorm:"column:id_category;type:uuid;default:gen_random_uuid();primaryKey"`
	UserID uuid.UUID `gorm:"column:user_id;type:uuid;not null"`
	Name   string    `gorm:"type:varchar(50);not null"`
}

func (Category) TableName() string {
	return "category"
}

func (c *CategoryRepositoryAdapter) CreateCategory(ctx context.Context, body *models.CategoryBody) error {
	category := Category{
		Name: body.Name,
	}

	result := c.db.WithContext(ctx).Create(&category)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (c *CategoryRepositoryAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	result := c.db.WithContext(ctx).Where("id = ?", id).Delete(&Category{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return models.ErrCategoryNotFound
	}

	return nil
}

func (c *CategoryRepositoryAdapter) GetAll(ctx context.Context, pageIndex, recordsPerPage int) ([]models.Category, error) {
	var categories []Category
	var modelCategories []models.Category

	offset := (pageIndex - 1) * recordsPerPage

	result := c.db.WithContext(ctx).
		Offset(offset).
		Limit(recordsPerPage).
		Find(&categories)

	if result.Error != nil {
		return nil, result.Error
	}

	for _, cat := range categories {
		modelCategories = append(modelCategories, models.Category{
			ID:   cat.ID,
			Name: cat.Name,
		})
	}

	return modelCategories, nil
}
