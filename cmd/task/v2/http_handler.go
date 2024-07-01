package task

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"todo-app-api/pkg/exception"
	"todo-app-api/pkg/middleware"
	"todo-app-api/pkg/response"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type TaskHTTPHandler struct {
	logger      *logrus.Logger
	validator   *validator.Validate
	taskUsecase TaskUsecase
}

func NewTaskHTTPHandler(logger *logrus.Logger, router *mux.Router, basicAuth middleware.RouteMiddleware, validator *validator.Validate, taskUsecase TaskUsecase) {
	handler := &TaskHTTPHandler{
		logger:      logger,
		validator:   validator,
		taskUsecase: taskUsecase,
	}
	router.HandleFunc("/todo/v2/task", basicAuth.Verify(handler.GetManyTasks)).Methods(http.MethodGet)
	router.HandleFunc("/todo/v2/task", basicAuth.Verify(handler.CreateTask)).Methods(http.MethodPost)
	router.HandleFunc("/todo/v2/task/{id}", basicAuth.Verify(handler.GetOneTask)).Methods(http.MethodGet)
	router.HandleFunc("/todo/v2/task/{id}", basicAuth.Verify(handler.UpdateTask)).Methods(http.MethodPut)
	router.HandleFunc("/todo/v2/task/attachment/{bucket}", basicAuth.Verify(handler.UploadAttachment)).Methods(http.MethodPost)
}

func (h TaskHTTPHandler) GetManyTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	qs := r.URL.Query()
	filter := GetManyTaskRequest{}

	if qs.Get("name") != "" {
		nameQs := qs.Get("name")
		filter.Name = &nameQs
	}

	resp := h.taskUsecase.GetManyTasks(ctx, filter)
	response.JSON(w, resp)
}

func (h TaskHTTPHandler) GetOneTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pathVariable := mux.Vars(r)
	taskId, _ := strconv.ParseInt(pathVariable["id"], 10, 64)
	resp := h.taskUsecase.GetOneTask(ctx, taskId)
	response.JSON(w, resp)
}

func (h TaskHTTPHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var resp response.Response
	var payload TaskRequest

	ctx := r.Context()

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		resp = response.NewErrorResponse(err, http.StatusUnprocessableEntity, nil, response.StatusInvalidPayload, err.Error())
		response.JSON(w, resp)
		return
	}

	if err := h.validateRequestBody(payload); err != nil {
		resp = response.NewErrorResponse(err, http.StatusBadRequest, nil, response.StatusInvalidPayload, err.Error())
		response.JSON(w, resp)
		return
	}

	resp = h.taskUsecase.CreateTask(ctx, payload)
	response.JSON(w, resp)
}

func (h TaskHTTPHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	var resp response.Response
	var payload TaskRequest

	pathVariable := mux.Vars(r)
	taskId, _ := strconv.ParseInt(pathVariable["id"], 10, 64)

	ctx := r.Context()

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		resp = response.NewErrorResponse(err, http.StatusUnprocessableEntity, nil, response.StatusInvalidPayload, err.Error())
		response.JSON(w, resp)
		return
	}

	if err := h.validateRequestBody(payload); err != nil {
		resp = response.NewErrorResponse(err, http.StatusBadRequest, nil, response.StatusInvalidPayload, err.Error())
		response.JSON(w, resp)
		return
	}

	resp = h.taskUsecase.UpdateTask(ctx, taskId, payload)
	response.JSON(w, resp)
}

func (h TaskHTTPHandler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	var resp response.Response
	var payload UploadAttachmentRequest

	pathVariable := mux.Vars(r)
	folderName := pathVariable["bucket"]

	ctx := r.Context()

	allowedFolderName := []string{"todo_attachment"}
	if err := h.validateFolderName(allowedFolderName, folderName); err != nil {
		resp = response.NewErrorResponse(err, http.StatusBadRequest, nil, response.StatusInvalidPayload, err.Error())
		response.JSON(w, resp)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.logger.WithContext(ctx).Error(err)
		resp := response.NewErrorResponse(exception.ErrBadRequest, http.StatusBadRequest, nil, response.StatusInvalidPayload, err.Error())
		response.JSON(w, resp)
		return
	}

	fileNameParam := r.Form.Get("filename")
	if fileNameParam == "" {
		resp := response.NewErrorResponse(exception.ErrBadRequest, http.StatusBadRequest, nil, response.StatusInvalidPayload, "Invalid file name parameter")
		response.JSON(w, resp)

		return
	}

	file, fileHeader, err := r.FormFile("attachment")
	if err != nil {
		h.logger.WithContext(ctx).Error(err)
		resp := response.NewErrorResponse(exception.ErrBadRequest, http.StatusBadRequest, nil, response.StatusInvalidPayload, err.Error())
		response.JSON(w, resp)
		return
	}

	defer file.Close()

	payload.Attachment.File = file
	payload.Attachment.FileName = fileHeader.Filename
	payload.Attachment.Size = fileHeader.Size
	payload.Attachment.FileExtension = filepath.Ext(fileHeader.Filename)
	payload.Attachment.FileNameParam = fileNameParam

	allowedExtension := []string{".jpeg", ".jpg", ".png"}
	if err := h.validateFileExtension(allowedExtension, payload.Attachment.FileExtension); err != nil {
		resp = response.NewErrorResponse(err, http.StatusBadRequest, nil, response.StatusInvalidPayload, err.Error())
		response.JSON(w, resp)
		return
	}

	if err := h.validateRequestBody(payload); err != nil {
		resp = response.NewErrorResponse(err, http.StatusBadRequest, nil, response.StatusInvalidPayload, err.Error())
		response.JSON(w, resp)
		return
	}

	resp = h.taskUsecase.UploadAttachment(ctx, folderName, payload)
	response.JSON(w, resp)
}

func (h TaskHTTPHandler) validateRequestBody(body interface{}) (err error) {
	err = h.validator.Struct(body)
	if err == nil {
		return
	}

	errorFields := err.(validator.ValidationErrors)
	errorField := errorFields[0]
	err = fmt.Errorf("invalid '%s' with value '%v'", errorField.Field(), errorField.Value())

	return
}

func (h TaskHTTPHandler) validateFolderName(elems []string, v string) (err error) {
	extension := strings.ToLower(v)
	for _, s := range elems {
		if extension == s {
			return
		}
	}

	err = fmt.Errorf("invalid bucket name '%s'", v)
	return
}

func (h TaskHTTPHandler) validateFileExtension(elems []string, v string) (err error) {
	extension := strings.ToLower(v)
	for _, s := range elems {
		if extension == s {
			return
		}
	}

	err = fmt.Errorf("invalid file extension '%s'", v)
	return
}
