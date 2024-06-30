package user

import (
	"context"
	"net/http"
	"time"
	"todo-app-api/pkg/exception"
	"todo-app-api/pkg/response"

	"github.com/sirupsen/logrus"
)

type UserUsecase interface {
	GetManyUsers(ctx context.Context) (resp response.Response)
	// GetOneTask(ctx context.Context, id int64) (resp response.Response)
	// CreateTask(ctx context.Context, taskRequest TaskRequest) (resp response.Response)
	// UpdateTask(ctx context.Context, id int64, taskRequest TaskRequest) (resp response.Response)
	// UploadAttachment(ctx context.Context, folderName string, payload UploadAttachmentRequest) (resp response.Response)
}

type userUsecase struct {
	logger         *logrus.Logger
	location       *time.Location
	userRepository UserRepository
}

func NewUserUsecase(logger *logrus.Logger, location *time.Location, userRepository UserRepository) UserUsecase {
	return &userUsecase{
		logger:         logger,
		location:       location,
		userRepository: userRepository,
	}
}

// GetManyUsers implements Usecase
func (u *userUsecase) GetManyUsers(ctx context.Context) (resp response.Response) {
	result, err := u.userRepository.FindManyUser(ctx)
	if err != nil {
		if err == exception.ErrNotFound {
			return response.NewErrorResponse(err, http.StatusNotFound, nil, response.StatNotFound, "")
		}
		u.logger.WithContext(ctx).Error(err)
		return response.NewErrorResponse(err, http.StatusInternalServerError, nil, response.StatUnexpectedError, "")
	}

	totalDataOnPage := len(result)
	usersResponse := make([]UserResponse, totalDataOnPage)
	for i, v := range result {
		var userResponse UserResponse
		userResponse.UUID = v.UUID
		userResponse.Name = v.Name
		userResponse.Email = v.Email
		userResponse.CreatedAt = v.CreatedAt

		usersResponse[i] = userResponse
	}

	return response.NewSuccessResponse(usersResponse, response.StatOK, "")
}
