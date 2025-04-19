package adapters

import (
	"context"
	"testing"
	mock_adapters "todolist/internal/adapters/mocks"
	"todolist/internal/models"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestUserAdapter_SignUp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_adapters.NewMockIUserRepository(ctrl)
	mockTokenHandler := mock_adapters.NewMockITokenHandler(ctrl)
	adapter := NewAuthService(mockRepo, mockTokenHandler, "test-key")

	tests := []struct {
		name          string
		candidate     *models.UserAuth
		mockSetup     func()
		expectedError error
	}{
		{
			name: "successful signup",
			candidate: &models.UserAuth{
				Name:     "testuser",
				Password: "password123",
			},
			mockSetup: func() {
				gomock.InOrder(
					mockRepo.EXPECT().
						CreateUser(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx context.Context, user *models.UserAuth) error {
							// Validate password was hashed
							err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("password123"))
							assert.Nil(t, err)
							return nil
						}),
				)
			},
			expectedError: nil,
		},
		{
			name: "empty username",
			candidate: &models.UserAuth{
				Name:     "",
				Password: "password123",
			},
			mockSetup:     func() {},
			expectedError: errors.New("Failed to login with empty login"),
		},
		{
			name: "empty password",
			candidate: &models.UserAuth{
				Name:     "testuser",
				Password: "",
			},
			mockSetup:     func() {},
			expectedError: errors.Errorf("Empty password for user with login %s", "testuser"),
		},
		{
			name: "repository error",
			candidate: &models.UserAuth{
				Name:     "testuser",
				Password: "password123",
			},
			mockSetup: func() {
				mockRepo.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			expectedError: errors.Wrapf(errors.New("database error"), "Failed to create user: %s", "testuser"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := adapter.SignUp(context.Background(), tt.candidate)

			if tt.expectedError != nil {
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestUserAdapter_SignIn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_adapters.NewMockIUserRepository(ctrl)
	mockTokenHandler := mock_adapters.NewMockITokenHandler(ctrl)
	adapter := NewAuthService(mockRepo, mockTokenHandler, "test-key")

	testUser := &models.User{
		ID:       uuid.New(),
		Name:     "testuser",
		Password: generateHash(t, "password123"),
	}

	tests := []struct {
		name          string
		candidate     *models.UserAuth
		mockSetup     func()
		expectedToken string
		expectedError error
	}{
		{
			name: "successful signin",
			candidate: &models.UserAuth{
				Name:     "testuser",
				Password: "password123",
			},
			mockSetup: func() {
				gomock.InOrder(
					mockRepo.EXPECT().
						GetUserByName(gomock.Any(), "testuser").
						Return(testUser, nil),
					mockTokenHandler.EXPECT().
						GenerateToken(*testUser, "test-key").
						Return("test-token", nil),
				)
			},
			expectedToken: "test-token",
			expectedError: nil,
		},
		{
			name: "invalid password",
			candidate: &models.UserAuth{
				Name:     "testuser",
				Password: "wrongpassword",
			},
			mockSetup: func() {
				mockRepo.EXPECT().
					GetUserByName(gomock.Any(), "testuser").
					Return(testUser, nil)
			},
			expectedToken: "",
			expectedError: errors.Wrapf(errors.New("crypto/bcrypt: hashedPassword is not the hash of the given password"), "Invalid password for user %s", "testuser"),
		},
		{
			name: "user not found",
			candidate: &models.UserAuth{
				Name:     "nonexistent",
				Password: "password123",
			},
			mockSetup: func() {
				mockRepo.EXPECT().
					GetUserByName(gomock.Any(), "nonexistent").
					Return(nil, errors.New("user not found"))
			},
			expectedToken: "",
			expectedError: errors.Wrapf(errors.New("user not found"), "Failed to get user %s", "nonexistent"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			token, err := adapter.SignIn(context.Background(), tt.candidate)

			if tt.expectedError != nil {
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.Equal(t, tt.expectedToken, token)
				assert.Nil(t, err)
			}
		})
	}
}

func TestUserAdapter_DeleteUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_adapters.NewMockIUserRepository(ctrl)
	mockTokenHandler := mock_adapters.NewMockITokenHandler(ctrl)
	adapter := NewAuthService(mockRepo, mockTokenHandler, "test-key")

	testID := uuid.New()

	tests := []struct {
		name          string
		userID        uuid.UUID
		mockSetup     func()
		expectedError error
	}{
		{
			name:   "successful deletion",
			userID: testID,
			mockSetup: func() {
				mockRepo.EXPECT().
					DeleteUser(gomock.Any(), testID).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "repository error",
			userID: testID,
			mockSetup: func() {
				mockRepo.EXPECT().
					DeleteUser(gomock.Any(), testID).
					Return(errors.New("deletion failed"))
			},
			expectedError: errors.Wrapf(errors.New("deletion failed"), "Failed to delete user with id %v", testID),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := adapter.DeleteUser(context.Background(), tt.userID)

			if tt.expectedError != nil {
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

// Вспомогательная функция для генерации хэша
func generateHash(t *testing.T, password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}
	return string(hash)
}
