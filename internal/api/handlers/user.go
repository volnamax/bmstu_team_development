package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"
	"todolist/internal/models"
	auth_utils "todolist/internal/pkg/authUtils"
	"todolist/internal/pkg/response"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type contextKey string // unexported base type

const (
	UserIDContextKey contextKey = "userID"
)

type UserInfo struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type Token struct {
	Token string `json:"token"`
}

type AuthProvider interface {
	SignIn(ctx context.Context, candidate *models.UserAuth) (tokenStr string, err error)
	SignUp(ctx context.Context, candidate *models.UserAuth) error
	CheckTaskOwnership(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) (bool, error)
	CheckCategoriesOwnership(ctx context.Context, userID uuid.UUID, categories []uuid.UUID) (bool, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
}

func FromUserInfo(userDTO UserInfo) *models.UserAuth {
	return &models.UserAuth{
		Name:     userDTO.Name,
		Password: userDTO.Password,
	}
}

// @Summary SignIn
// @Tags user
// @Description SignIn
// @ID sign-in
// @Accept  json
// @Produce  json
// @Param input body UserInfo true "user's name and password"
// @Success 200 {object} Token
// @Failure 400,404 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure default {object} response.Response
// @Router /api/v1/sign-in [post]
func SignIn(authProvider AuthProvider, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := zerolog.Ctx(r.Context())

		logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("SignIn: started authentication process")

		var req UserInfo
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			logger.Warn().
				Err(err).
				Msg("SignIn: failed to decode request body")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		logger.Debug().
			Str("username", req.Name).
			Msg("SignIn: processing authentication request")

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		tokenStr, err := authProvider.SignIn(ctx, FromUserInfo(req))
		if err != nil {
			if errors.Is(err, auth_utils.ErrInvalidToken) {
				logger.Warn().
					Str("username", req.Name).
					Err(err).
					Msg("SignIn: invalid credentials provided")
				render.Status(r, http.StatusBadRequest)
			} else if errors.Is(err, models.ErrUserNotFound) {
				logger.Warn().
					Str("username", req.Name).
					Err(err).
					Msg("SignIn: user not found")
				render.Status(r, http.StatusNotFound)
			} else {
				logger.Error().
					Str("username", req.Name).
					Err(err).
					Msg("SignIn: authentication failed")
				render.Status(r, http.StatusInternalServerError)
			}
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		logger.Info().
			Str("username", req.Name).
			Msg("SignIn: successfully authenticated user")
		render.JSON(w, r, Token{Token: tokenStr})
	}
}

// @Summary SignUp
// @Tags user
// @Description SignUp
// @ID sign-up
// @Accept  json
// @Produce  json
// @Param input body UserInfo true "user's name and password"
// @Success 200 {object} Token
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure default {object} response.Response
// @Router /api/v1/sign-up [post]
func SignUp(authProvider AuthProvider, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := zerolog.Ctx(r.Context())

		logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("SignUp: started registration process")

		var req UserInfo
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			logger.Warn().
				Err(err).
				Msg("SignUp: failed to decode request body")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		logger.Debug().
			Str("username", req.Name).
			Msg("SignUp: processing registration request")

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err = authProvider.SignUp(ctx, FromUserInfo(req))
		if err != nil {
			logger.Error().
				Str("username", req.Name).
				Err(err).
				Msg("SignUp: failed to register user")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		logger.Info().
			Str("username", req.Name).
			Msg("SignUp: successfully registered new user")
		render.Status(r, http.StatusOK)
	}
}

// @Summary DeleteUser
// @Security ApiKeyAuth
// @Tags user
// @Description delete user
// @ID delete-user
// @Accept  json
// @Produce  json
// @Success 200
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure default {object} response.Response
// @Router /api/v1/user [delete]
func DeleteUser(authProvider AuthProvider, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := zerolog.Ctx(r.Context())

		logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("DeleteUser: started processing request")

		userID, ok := r.Context().Value(UserIDContextKey).(string)
		if !ok {
			logger.Warn().
				Msg("DeleteUser: missing userID in context")
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, response.Error("Missing userID"))
			return
		}

		logger.Debug().
			Str("user_id", userID).
			Msg("DeleteUser: processing user deletion")

		userUUID, err := uuid.Parse(userID)
		if err != nil {
			logger.Warn().
				Str("user_id", userID).
				Err(err).
				Msg("DeleteUser: malformed user ID format")
			render.JSON(w, r, response.Error("Invalid userID"))
			render.Status(r, http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err = authProvider.DeleteUser(ctx, userUUID)
		if err != nil {
			logger.Error().
				Str("user_id", userUUID.String()).
				Err(err).
				Msg("DeleteUser: failed to delete user")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		logger.Info().
			Str("user_id", userUUID.String()).
			Msg("DeleteUser: successfully deleted user")
		render.Status(r, http.StatusOK)
	}
}
