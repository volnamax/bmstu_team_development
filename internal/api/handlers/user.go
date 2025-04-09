package handlers

import "net/http"

type UserInfo struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type Token struct {
	Token string `json:"token"`
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
func SignIn() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

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
// @Failure 400,404 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure default {object} response.Response
// @Router /api/v1/sign-up [post]
func SignUp() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
