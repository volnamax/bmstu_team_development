package repository

import (
	"context"
	"fmt"
	"todolist/internal/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type User struct {
	ID       uuid.UUID `gorm:"primaryKey;column:id_user;type:uuid;default:gen_random_uuid()"`
	Name     string    `gorm:"unique;column:user_name"`
	Password string    `gorm:"column:password_hash"`
}

func ToDaUser(user models.UserAuth) User {
	return User{
		Name:     user.Name,
		Password: user.Password,
	}
}

func FromDaUser(user User) models.User {
	return models.User{
		ID:       user.ID,
		Name:     user.Name,
		Password: user.Password,
	}
}

type UserRepositoryAdapter struct {
	db *gorm.DB
}

func NewUserRepositoryAdapter(srcDB *gorm.DB) *UserRepositoryAdapter {
	return &UserRepositoryAdapter{
		db: srcDB,
	}
}

func (repo *UserRepositoryAdapter) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var userDA User
	userDA.ID = id

	tx := repo.db.WithContext(ctx).First(&userDA)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, models.ErrUserNotFound
		}
		return nil, errors.Wrap(tx.Error, "error getting user by ID")
	}

	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context error during user lookup")
	}

	user := FromDaUser(userDA)
	return &user, nil
}

func (repo *UserRepositoryAdapter) GetUserByName(ctx context.Context, name string) (*models.User, error) {
	var userDA User

	tx := repo.db.WithContext(ctx).Where("user_name = ?", name).First(&userDA)
	fmt.Print(userDA)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, models.ErrUserNotFound
		}
		return nil, errors.Wrap(tx.Error, "error getting user by name")
	}

	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context error during user lookup")
	}

	user := FromDaUser(userDA)
	return &user, nil
}

func (repo *UserRepositoryAdapter) CreateUser(ctx context.Context, user *models.UserAuth) error {
	userDa := ToDaUser(*user)

	tx := repo.db.WithContext(ctx).Create(&userDa)
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "error creating user")
	}

	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context error during user creation")
	}

	return nil
}

func (repo *UserRepositoryAdapter) CheckTaskOwnership(ctx context.Context, userID, taskID uuid.UUID) (bool, error) {
	var isOwned bool

	tx := repo.db.WithContext(ctx).
		Raw("SELECT EXISTS(SELECT 1 FROM task WHERE id_task = ? AND user_id = ?)",
			taskID, userID).
		Scan(&isOwned)

	if tx.Error != nil {
		return false, errors.Wrap(tx.Error, "failed to verify task ownership")
	}

	if err := ctx.Err(); err != nil {
		return false, errors.Wrap(err, "context error during ownership check")
	}

	return isOwned, nil
}

func (repo *UserRepositoryAdapter) CheckCategoriesOwnership(ctx context.Context, userID uuid.UUID, categories []uuid.UUID) (bool, error) {
	if len(categories) == 0 {
		return true, nil
	}

	// Convert UUID slice to string slice
	categoryStrings := make([]string, len(categories))
	for i, cat := range categories {
		categoryStrings[i] = cat.String()
	}

	var allOwned bool

	tx := repo.db.WithContext(ctx).Raw(`
        SELECT  NOT EXISTS (
            SELECT 1 FROM category 
            WHERE id_category = ANY(?::uuid[]) 
            AND user_id != ?
        )`, pq.Array(categoryStrings), userID).Scan(&allOwned)

	if tx.Error != nil {
		return false, errors.Wrap(tx.Error, "failed raw ownership check")
	}

	if err := ctx.Err(); err != nil {
		return false, errors.Wrap(err, "context error during categories check")
	}

	return allOwned, nil
}

func (repo *UserRepositoryAdapter) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	tx := repo.db.WithContext(ctx).Delete(&User{}, "id_user = ?", userID)
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "error deleting user")
	}

	if tx.RowsAffected == 0 {
		return models.ErrUserNotFound
	}

	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context error during user deletion")
	}

	return nil
}
