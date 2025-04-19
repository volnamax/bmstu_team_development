package repository

import (
	"context"

	"todolist/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Task struct {
	ID          uuid.UUID  `gorm:"column:id_task;type:uuid;default:gen_random_uuid();primaryKey"`
	UserID      uuid.UUID  `gorm:"column:user_id;type:uuid;not null"`
	Title       string     `gorm:"type:varchar(128);not null"`
	Description string     `gorm:"type:varchar(1000)"`
	IsDone      bool       `gorm:"column:is_done;default:false"`
	Categories  []Category `gorm:"many2many:task_category;joinForeignKey:TaskID;JoinReferences:CategoryID"`
}

func (Task) TableName() string {
	return "task"
}

type GormTaskRepository struct {
	db *gorm.DB
}

func NewGormTaskRepository(db *gorm.DB) *GormTaskRepository {
	return &GormTaskRepository{db: db}
}

func (r *GormTaskRepository) CreateTask(ctx context.Context, userId uuid.UUID, body *models.TaskBody, categoryIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		task := Task{
			UserID:      userId,
			Title:       body.Title,
			Description: body.Description,
			IsDone:      false,
		}

		if err := tx.Create(&task).Error; err != nil {
			return err
		}

		if len(categoryIDs) > 0 {
			var categories []Category
			if err := tx.
				Where("id_category IN ?", categoryIDs).
				Find(&categories).Error; err != nil {
				return err
			}
			if err := tx.Model(&task).Association("Categories").Replace(&categories); err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *GormTaskRepository) Update(ctx context.Context, id uuid.UUID, body *models.TaskBody, categoryIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var task Task
		if err := tx.First(&task, "id_task = ?", id).Error; err != nil {
			return err
		}

		task.Title = body.Title
		task.Description = body.Description

		if err := tx.Save(&task).Error; err != nil {
			return err
		}

		if categoryIDs != nil {
			var categories []Category
			if err := tx.
				Where("id_category IN ?", categoryIDs).
				Find(&categories).Error; err != nil {
				return err
			}
			if err := tx.Model(&task).Association("Categories").Replace(&categories); err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *GormTaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.TaskFullInfo, error) {
	var task Task
	if err := r.db.WithContext(ctx).
		Preload("Categories").
		First(&task, "id_task = ?", id).Error; err != nil {
		return nil, err
	}

	categoryNames := make([]models.Category, len(task.Categories))
	for i, cat := range task.Categories {
		categoryNames[i] = models.Category{
			ID:   cat.ID,
			Name: cat.Name,
		}
	}

	return &models.TaskFullInfo{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		IsDone:      task.IsDone,
		Categories:  categoryNames,
	}, nil
}

func (r *GormTaskRepository) GetAll(ctx context.Context, userId uuid.UUID, pageIndex, recordsPerPage int) ([]models.TaskShortInfo, error) {
	var tasks []Task
	offset := (pageIndex - 1) * recordsPerPage

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userId).
		Order("title ASC").
		Limit(recordsPerPage).
		Offset(offset).
		Find(&tasks).Error

	if err != nil {
		return nil, err
	}

	result := make([]models.TaskShortInfo, len(tasks))
	for i, task := range tasks {
		result[i] = models.TaskShortInfo{
			ID:     task.ID,
			Title:  task.Title,
			IsDone: task.IsDone,
		}
	}

	return result, nil
}

func (r *GormTaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&Task{}, "id_task = ?", id).Error; err != nil {
		return err
	}
	return nil
}

func (r *GormTaskRepository) ToggleDone(ctx context.Context, id uuid.UUID) error {
	var task Task
	if err := r.db.WithContext(ctx).First(&task, "id_task = ?", id).Error; err != nil {
		return err
	}

	task.IsDone = !task.IsDone

	if err := r.db.WithContext(ctx).Save(&task).Error; err != nil {
		return err
	}

	return nil
}
