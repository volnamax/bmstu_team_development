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
)

func TestCategoryAdapter_CreateCategory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_adapters.NewMockCategoryRepository(ctrl)
	adapter := NewCategoryAdapter(mockRepo)

	tests := []struct {
		name          string
		ctx           context.Context
		body          *models.CategoryBody
		mockSetup     func()
		expectedError error
	}{
		{
			name: "successful creation",
			ctx:  context.Background(),
			body: &models.CategoryBody{Name: "Test Category"},
			mockSetup: func() {
				mockRepo.EXPECT().CreateCategory(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "repository error",
			ctx:  context.Background(),
			body: &models.CategoryBody{Name: "Test Category"},
			mockSetup: func() {
				mockRepo.EXPECT().CreateCategory(gomock.Any(), gomock.Any()).Return(errors.New("db error"))
			},
			expectedError: errors.New("failed to create category"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := adapter.CreateCategory(tt.ctx, tt.body)

			if tt.expectedError != nil {
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestCategoryAdapter_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_adapters.NewMockCategoryRepository(ctrl)
	adapter := NewCategoryAdapter(mockRepo)

	testID := uuid.New()

	tests := []struct {
		name          string
		id            uuid.UUID
		mockSetup     func()
		expectedError error
	}{
		{
			name: "successful deletion",
			id:   testID,
			mockSetup: func() {
				mockRepo.EXPECT().Delete(gomock.Any(), testID).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "repository error",
			id:   testID,
			mockSetup: func() {
				mockRepo.EXPECT().Delete(gomock.Any(), testID).Return(errors.New("db error"))
			},
			expectedError: errors.New("failed to delete category"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := adapter.Delete(context.Background(), tt.id)

			if tt.expectedError != nil {
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestCategoryAdapter_GetAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_adapters.NewMockCategoryRepository(ctrl)
	adapter := NewCategoryAdapter(mockRepo)

	testUserID := uuid.New()
	testCategories := []models.Category{{ID: uuid.New(), Name: "Test 1"}, {ID: uuid.New(), Name: "Test 2"}}

	tests := []struct {
		name           string
		pageIndex      int
		recordsPerPage int
		userID         uuid.UUID
		mockSetup      func()
		expectedOutput []models.Category
		expectedError  error
	}{
		{
			name:           "successful fetch",
			pageIndex:      1,
			recordsPerPage: 10,
			userID:         testUserID,
			mockSetup: func() {
				mockRepo.EXPECT().GetAll(gomock.Any(), 1, 10, testUserID).Return(testCategories, nil)
			},
			expectedOutput: testCategories,
			expectedError:  nil,
		},
		{
			name:           "repository error",
			pageIndex:      1,
			recordsPerPage: 10,
			userID:         testUserID,
			mockSetup: func() {
				mockRepo.EXPECT().GetAll(gomock.Any(), 1, 10, testUserID).Return(nil, errors.New("db error"))
			},
			expectedOutput: nil,
			expectedError:  errors.New("failed to get all categories"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := adapter.GetAll(context.Background(), tt.pageIndex, tt.recordsPerPage, tt.userID)

			if tt.expectedError != nil {
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.Equal(t, tt.expectedOutput, result)
			}
		})
	}
}
