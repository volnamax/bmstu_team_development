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

func TestTaskAdapter_CreateTask(t *testing.T) {
	type mockBehavior func(r *mock_adapters.MockTaskRepository, ctx context.Context, userID uuid.UUID, body *models.TaskBody, catIDs []uuid.UUID)

	testTable := []struct {
		name        string
		userID      uuid.UUID
		body        *models.TaskBody
		categoryIDs []uuid.UUID
		mock        mockBehavior
		expectedErr error
	}{
		{
			name:        "success",
			userID:      uuid.New(),
			body:        &models.TaskBody{Title: "task"},
			categoryIDs: []uuid.UUID{uuid.New()},
			mock: func(r *mock_adapters.MockTaskRepository, ctx context.Context, userID uuid.UUID, body *models.TaskBody, catIDs []uuid.UUID) {
				r.EXPECT().CreateTask(ctx, userID, body, catIDs).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "repository error",
			userID:      uuid.New(),
			body:        &models.TaskBody{Title: "fail"},
			categoryIDs: []uuid.UUID{uuid.New()},
			mock: func(r *mock_adapters.MockTaskRepository, ctx context.Context, userID uuid.UUID, body *models.TaskBody, catIDs []uuid.UUID) {
				r.EXPECT().CreateTask(ctx, userID, body, catIDs).Return(errors.New("repo error"))
			},
			expectedErr: errors.Wrap(errors.New("repo error"), "failed to create task"),
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_adapters.NewMockTaskRepository(ctrl)
			ctx := context.Background()
			tc.mock(mockRepo, ctx, tc.userID, tc.body, tc.categoryIDs)

			adapter := NewTaskAdapter(mockRepo)
			err := adapter.CreateTask(ctx, tc.userID, tc.body, tc.categoryIDs)

			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskAdapter_Update(t *testing.T) {
	type mockBehavior func(r *mock_adapters.MockTaskRepository, ctx context.Context, taskID uuid.UUID, body *models.TaskBody, catIDs []uuid.UUID)

	testTable := []struct {
		name        string
		taskID      uuid.UUID
		body        *models.TaskBody
		categoryIDs []uuid.UUID
		mock        mockBehavior
		expectedErr error
	}{
		{
			name:        "successful update",
			taskID:      uuid.New(),
			body:        &models.TaskBody{Title: "updated title"},
			categoryIDs: []uuid.UUID{uuid.New()},
			mock: func(r *mock_adapters.MockTaskRepository, ctx context.Context, taskID uuid.UUID, body *models.TaskBody, catIDs []uuid.UUID) {
				r.EXPECT().Update(ctx, taskID, body, catIDs).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "repository error",
			taskID:      uuid.New(),
			body:        &models.TaskBody{Title: "will fail"},
			categoryIDs: []uuid.UUID{uuid.New()},
			mock: func(r *mock_adapters.MockTaskRepository, ctx context.Context, taskID uuid.UUID, body *models.TaskBody, catIDs []uuid.UUID) {
				r.EXPECT().Update(ctx, taskID, body, catIDs).Return(errors.New("update failed"))
			},
			expectedErr: errors.Wrapf(errors.New("update failed"), "failed to update task with id: %s", taskID),
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_adapters.NewMockTaskRepository(ctrl)
			ctx := context.Background()
			tc.mock(mockRepo, ctx, tc.taskID, tc.body, tc.categoryIDs)

			adapter := NewTaskAdapter(mockRepo)
			err := adapter.Update(ctx, tc.taskID, tc.body, tc.categoryIDs)

			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskAdapter_GetByID(t *testing.T) {
	type mockBehavior func(r *mock_adapters.MockTaskRepository, ctx context.Context, taskID uuid.UUID)

	testTable := []struct {
		name         string
		taskID       uuid.UUID
		mock         mockBehavior
		expectedTask *models.TaskFullInfo
		expectedErr  error
	}{
		{
			name:   "success",
			taskID: uuid.New(),
			mock: func(r *mock_adapters.MockTaskRepository, ctx context.Context, taskID uuid.UUID) {
				task := &models.TaskFullInfo{
					ID:    taskID,
					Title: "Test Task",
				}
				r.EXPECT().GetByID(ctx, taskID).Return(task, nil)
			},
			expectedTask: &models.TaskFullInfo{
				ID:    uuid.Nil, // заменим позже в тесте
				Title: "Test Task",
			},
			expectedErr: nil,
		},
		{
			name:   "repository error",
			taskID: uuid.New(),
			mock: func(r *mock_adapters.MockTaskRepository, ctx context.Context, taskID uuid.UUID) {
				r.EXPECT().GetByID(ctx, taskID).Return(nil, errors.New("not found"))
			},
			expectedTask: nil,
			expectedErr:  errors.Wrapf(errors.New("not found"), "failed to get task by id: %s", uuid.Nil),
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_adapters.NewMockTaskRepository(ctrl)
			ctx := context.Background()

			// Прокидываем taskID в expectedErr и expectedTask, где надо
			if tc.expectedTask != nil {
				tc.expectedTask.ID = tc.taskID
			}
			if tc.expectedErr != nil {
				tc.expectedErr = errors.Wrapf(errors.Unwrap(tc.expectedErr), "failed to get task by id: %s", tc.taskID)
			}

			tc.mock(mockRepo, ctx, tc.taskID)

			adapter := NewTaskAdapter(mockRepo)
			task, err := adapter.GetByID(ctx, tc.taskID)

			assert.Equal(t, tc.expectedTask, task)

			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskAdapter_GetAll(t *testing.T) {
	type mockBehavior func(r *mock_adapters.MockTaskRepository, ctx context.Context, userID uuid.UUID, pageIndex, recordsPerPage int)

	testTable := []struct {
		name           string
		userID         uuid.UUID
		pageIndex      int
		recordsPerPage int
		mock           mockBehavior
		expectedTasks  []models.TaskShortInfo
		expectedErr    error
	}{
		{
			name:           "success",
			userID:         uuid.New(),
			pageIndex:      1,
			recordsPerPage: 10,
			mock: func(r *mock_adapters.MockTaskRepository, ctx context.Context, userID uuid.UUID, pageIndex, recordsPerPage int) {
				tasks := []models.TaskShortInfo{
					{ID: uuid.New(), Title: "Task 1"},
					{ID: uuid.New(), Title: "Task 2"},
				}
				r.EXPECT().GetAll(ctx, userID, pageIndex, recordsPerPage).Return(tasks, nil)
			},
			expectedTasks: []models.TaskShortInfo{
				{Title: "Task 1"},
				{Title: "Task 2"},
			},
			expectedErr: nil,
		},
		{
			name:           "repository error",
			userID:         uuid.New(),
			pageIndex:      2,
			recordsPerPage: 5,
			mock: func(r *mock_adapters.MockTaskRepository, ctx context.Context, userID uuid.UUID, pageIndex, recordsPerPage int) {
				r.EXPECT().GetAll(ctx, userID, pageIndex, recordsPerPage).Return(nil, errors.New("db error"))
			},
			expectedTasks: nil,
			expectedErr:   errors.Wrap(errors.New("db error"), "failed to get all tasks"),
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_adapters.NewMockTaskRepository(ctrl)
			ctx := context.Background()
			tc.mock(mockRepo, ctx, tc.userID, tc.pageIndex, tc.recordsPerPage)

			adapter := NewTaskAdapter(mockRepo)
			tasks, err := adapter.GetAll(ctx, tc.userID, tc.pageIndex, tc.recordsPerPage)

			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
				assert.Nil(t, tasks)
			} else {
				assert.NoError(t, err)
				// Сравниваем только по Title, чтобы не заморачиваться с uuid
				for i := range tc.expectedTasks {
					assert.Equal(t, tc.expectedTasks[i].Title, tasks[i].Title)
				}
			}
		})
	}
}

func TestTaskAdapter_Delete(t *testing.T) {
	type mockBehavior func(r *mock_adapters.MockTaskRepository, ctx context.Context, taskID uuid.UUID)

	testTable := []struct {
		name        string
		taskID      uuid.UUID
		mock        mockBehavior
		expectedErr error
	}{
		{
			name:   "success",
			taskID: uuid.New(),
			mock: func(r *mock_adapters.MockTaskRepository, ctx context.Context, taskID uuid.UUID) {
				r.EXPECT().Delete(ctx, taskID).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:   "repository error",
			taskID: uuid.New(),
			mock: func(r *mock_adapters.MockTaskRepository, ctx context.Context, taskID uuid.UUID) {
				r.EXPECT().Delete(ctx, taskID).Return(errors.New("delete failed"))
			},
			expectedErr: errors.Wrapf(errors.New("delete failed"), "failed to delete task with id: %s", uuid.Nil), // заменим позже
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_adapters.NewMockTaskRepository(ctrl)
			ctx := context.Background()

			if tc.expectedErr != nil {
				tc.expectedErr = errors.Wrapf(errors.Unwrap(tc.expectedErr), "failed to delete task with id: %s", tc.taskID)
			}

			tc.mock(mockRepo, ctx, tc.taskID)

			adapter := NewTaskAdapter(mockRepo)
			err := adapter.Delete(ctx, tc.taskID)

			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskAdapter_ToggleDone(t *testing.T) {
	type mockBehavior func(r *mock_adapters.MockTaskRepository, ctx context.Context, taskID uuid.UUID)

	testTable := []struct {
		name        string
		taskID      uuid.UUID
		mock        mockBehavior
		expectedErr error
	}{
		{
			name:   "success",
			taskID: uuid.New(),
			mock: func(r *mock_adapters.MockTaskRepository, ctx context.Context, taskID uuid.UUID) {
				r.EXPECT().ToggleDone(ctx, taskID).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:   "repository error",
			taskID: uuid.New(),
			mock: func(r *mock_adapters.MockTaskRepository, ctx context.Context, taskID uuid.UUID) {
				r.EXPECT().ToggleDone(ctx, taskID).Return(errors.New("toggle failed"))
			},
			expectedErr: errors.Wrapf(errors.New("toggle failed"), "failed to toggle task done status with id: %s", uuid.Nil), // заменим позже
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_adapters.NewMockTaskRepository(ctrl)
			ctx := context.Background()

			if tc.expectedErr != nil {
				tc.expectedErr = errors.Wrapf(errors.Unwrap(tc.expectedErr), "failed to toggle task done status with id: %s", tc.taskID)
			}

			tc.mock(mockRepo, ctx, tc.taskID)

			adapter := NewTaskAdapter(mockRepo)
			err := adapter.ToggleDone(ctx, tc.taskID)

			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
