package user

import (
	"net/http"
	"todo-app-api/pkg/middleware"
	"todo-app-api/pkg/response"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type UserHTTPHandler struct {
	logger      *logrus.Logger
	validator   *validator.Validate
	userUsecase UserUsecase
}

func NewUserHTTPHandler(logger *logrus.Logger, router *mux.Router, basicAuth middleware.RouteMiddleware, validator *validator.Validate, userUsecase UserUsecase) {
	handler := &UserHTTPHandler{
		logger:      logger,
		validator:   validator,
		userUsecase: userUsecase,
	}
	router.HandleFunc("/api/v1/user", basicAuth.Verify(handler.GetManyUsers)).Methods(http.MethodGet)
	// router.HandleFunc("/todo/v1/task", basicAuth.Verify(handler.CreateTask)).Methods(http.MethodPost)
	// router.HandleFunc("/todo/v1/task/{id}", basicAuth.Verify(handler.GetOneTask)).Methods(http.MethodGet)
	// router.HandleFunc("/todo/v1/task/{id}", basicAuth.Verify(handler.UpdateTask)).Methods(http.MethodPut)
	// router.HandleFunc("/todo/v1/task/attachment/{bucket}", basicAuth.Verify(handler.UploadAttachment)).Methods(http.MethodPost)
}

func (h UserHTTPHandler) GetManyUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	resp := h.userUsecase.GetManyUsers(ctx)
	response.JSON(w, resp)
}
