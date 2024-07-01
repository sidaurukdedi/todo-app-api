package task

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"todo-app-api/entity"
	"todo-app-api/pkg/exception"
	"todo-app-api/pkg/response"
	"todo-app-api/pkg/storage"

	"github.com/sirupsen/logrus"
)

type TaskUsecase interface {
	GetManyTasks(ctx context.Context, filter GetManyTaskRequest) (resp response.Response)
	GetOneTask(ctx context.Context, id int64) (resp response.Response)
	CreateTask(ctx context.Context, taskRequest TaskRequest) (resp response.Response)
	UpdateTask(ctx context.Context, id int64, taskRequest TaskRequest) (resp response.Response)
	UploadAttachment(ctx context.Context, folderName string, payload UploadAttachmentRequest) (resp response.Response)
}

type taskUsecase struct {
	logger         *logrus.Logger
	location       *time.Location
	storage        storage.Storage
	taskRepository TaskRepository
}

func NewTaskUsecase(logger *logrus.Logger, location *time.Location, storage storage.Storage, taskRepository TaskRepository) TaskUsecase {
	return &taskUsecase{
		logger:         logger,
		location:       location,
		storage:        storage,
		taskRepository: taskRepository,
	}
}

// GetManyTasks implements Usecase
func (u *taskUsecase) GetManyTasks(ctx context.Context, filter GetManyTaskRequest) (resp response.Response) {
	result, err := u.taskRepository.FindMany(ctx, filter)
	if err != nil {
		if err == exception.ErrNotFound {
			return response.NewErrorResponse(err, http.StatusNotFound, nil, response.StatNotFound, "")
		}
		u.logger.WithContext(ctx).Error(err)
		return response.NewErrorResponse(err, http.StatusInternalServerError, nil, response.StatUnexpectedError, "")
	}

	totalDataOnPage := len(result)
	tasksResponse := make([]TaskResponse, totalDataOnPage)
	for i, v := range result {
		var taskResponse TaskResponse
		taskResponse.ID = v.ID
		taskResponse.Name = v.Name
		taskResponse.Description = &v.Description
		taskResponse.Status = &v.Status
		taskResponse.Attachment = v.Attachment
		taskResponse.CreatedAt = v.CreatedAt
		taskResponse.UpdatedAt = v.UpdatedAt

		tasksResponse[i] = taskResponse
	}

	return response.NewSuccessResponse(tasksResponse, response.StatOK, "")
}

// GetOneTask implements Usecase
func (u *taskUsecase) GetOneTask(ctx context.Context, id int64) (resp response.Response) {
	result, err := u.taskRepository.FindOneById(ctx, id)
	if err != nil {
		if err == exception.ErrNotFound {
			return response.NewErrorResponse(err, http.StatusNotFound, nil, response.StatNotFound, "")
		}
		u.logger.WithContext(ctx).Error(err)
		return response.NewErrorResponse(err, http.StatusInternalServerError, nil, response.StatUnexpectedError, "")
	}

	taskResponse := TaskResponse{
		ID:          result.ID,
		Name:        result.Name,
		Description: &result.Description,
		Status:      &result.Status,
		Attachment:  result.Attachment,
		CreatedAt:   result.CreatedAt,
		UpdatedAt:   result.UpdatedAt,
	}

	return response.NewSuccessResponse(taskResponse, response.StatOK, "")
}

// CreateTask implements Usecase
func (u *taskUsecase) CreateTask(ctx context.Context, taskRequest TaskRequest) (resp response.Response) {
	taskStatus := entity.TaskStatusInitiate
	createdAt := time.Now().In(u.location)

	taskRequest.Status = &taskStatus
	taskRequest.CreatedAt = createdAt

	taskId, err := u.taskRepository.Save(ctx, taskRequest, nil)
	if err != nil {
		return response.NewErrorResponse(exception.ErrInternalServer, http.StatusInternalServerError, nil, response.StatUnexpectedError, "")
	}

	taskResponse := TaskResponse{
		ID:          taskId,
		Name:        taskRequest.Name,
		Description: taskRequest.Description,
		Attachment:  taskRequest.Attachment,
		CreatedAt:   createdAt,
	}

	return response.NewSuccessResponse(taskResponse, response.StatOK, "")
}

// UpdateTask implements Usecase
func (u *taskUsecase) UpdateTask(ctx context.Context, id int64, taskRequest TaskRequest) (resp response.Response) {
	task, err := u.taskRepository.FindOneById(ctx, id)
	if err != nil {
		if err == exception.ErrNotFound {
			return response.NewErrorResponse(err, http.StatusNotFound, nil, response.StatNotFound, "")
		}
		u.logger.WithContext(ctx).Error(err)
		return response.NewErrorResponse(err, http.StatusInternalServerError, nil, response.StatUnexpectedError, "")
	}

	updatedAt := time.Now().In(u.location)
	taskRequest.UpdatedAt = &updatedAt

	err = u.taskRepository.UpdateById(ctx, id, taskRequest, nil)
	if err != nil {
		return response.NewErrorResponse(exception.ErrInternalServer, http.StatusInternalServerError, nil, response.StatUnexpectedError, "")
	}

	taskResponse := TaskResponse{
		ID:          id,
		Name:        taskRequest.Name,
		Description: taskRequest.Description,
		Status:      taskRequest.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   &updatedAt,
	}

	return response.NewSuccessResponse(taskResponse, response.StatOK, "")
}

// UploadAttachment implements Usecase
func (u *taskUsecase) UploadAttachment(ctx context.Context, folderName string, payload UploadAttachmentRequest) (resp response.Response) {
	var attachment entity.Attachment
	var url string

	bucketName := "image-wreg"
	fileName := fmt.Sprintf("wr/%s/%s%s", folderName, payload.Attachment.FileNameParam, payload.Attachment.FileExtension)

	err := u.storage.PutObject(context.Background(), bucketName, fileName, payload.Attachment.File, "image/png", nil)
	if err != nil {
		u.logger.WithContext(ctx).Error(err)
		return response.NewErrorResponse(exception.ErrInternalServer, http.StatusInternalServerError, nil, response.StatUnexpectedError, "")
	}

	host := "https://storage.googleapis.com/"
	url = fmt.Sprintf("%s%s/%s", host, bucketName, fileName)
	attachment.ImageURL = url

	return response.NewSuccessResponse(attachment, response.StatOK, "")
}
