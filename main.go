package main

import (
	"context"
	"database/sql"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	taskV1 "todo-app-api/cmd/task/v1"
	taskV2 "todo-app-api/cmd/task/v2"
	"todo-app-api/cmd/user/v1"
	"todo-app-api/configs"
	"todo-app-api/pkg/hook"
	"todo-app-api/pkg/middleware"
	"todo-app-api/pkg/response"
	"todo-app-api/server"

	gcs "cloud.google.com/go/storage"
	"github.com/go-playground/validator/v10"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload" // for development
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	ddlogrus "gopkg.in/DataDog/dd-trace-go.v1/contrib/sirupsen/logrus"

	s "todo-app-api/pkg/storage"
)

var (
	// tracer       *apm.Tracer
	cfg          *configs.Config
	indexMessage string = "Application is running properly"
)

func init() {
	// tracer = apm.DefaultTracer
	cfg = configs.Load()
}

func main() {
	logger := logrus.New()
	logger.SetFormatter(cfg.Logger.Formatter)
	logger.SetReportCaller(true)
	logger.AddHook(&ddlogrus.DDContextLogHook{})
	logger.AddHook(hook.NewStdoutLoggerHook(logrus.New(), cfg.Logger.Formatter))

	// set mariadb read only object
	dbReadOnly, err := sql.Open(cfg.MariadbReadOnly.Driver, cfg.MariadbReadOnly.DSN)
	if err != nil {
		logger.Fatal(err)
	}
	if err := dbReadOnly.Ping(); err != nil {
		logger.Fatal(err)
	}
	dbReadOnly.SetConnMaxLifetime(time.Minute * 3)
	dbReadOnly.SetMaxOpenConns(cfg.MariadbReadOnly.MaxOpenConnections)
	dbReadOnly.SetMaxIdleConns(cfg.MariadbReadOnly.MaxIdleConnections)

	// set mariadb read write object
	dbReadWrite, err := sql.Open(cfg.MariadbReadWrite.Driver, cfg.MariadbReadWrite.DSN)
	if err != nil {
		logger.Fatal(err)
	}
	if err := dbReadWrite.Ping(); err != nil {
		logger.Fatal(err)
	}
	dbReadWrite.SetConnMaxLifetime(time.Minute * 3)
	dbReadWrite.SetMaxOpenConns(cfg.MariadbReadWrite.MaxOpenConnections)
	dbReadWrite.SetMaxIdleConns(cfg.MariadbReadWrite.MaxIdleConnections)

	router := mux.NewRouter()
	router.HandleFunc("/todo", index)

	basicAuthMiddleware := middleware.NewBasicAuth(cfg.BasicAuth.Username, cfg.BasicAuth.Password)

	// set google cloud storage
	credentials, _ := ioutil.ReadFile("./secret/gcp_credential.json")
	gcsclient, _ := gcs.NewClient(context.Background(), option.WithCredentialsJSON(credentials))
	gcs := s.NewGCSAdapter(gcsclient, cfg.GCPStorage.AccessID, string(cfg.GCPStorage.PrivateKey))

	// set validator
	validator := validator.New()
	// validator.RegisterTagNameFunc(customvalidator.SetTagName)
	// validator.RegisterValidation("default-name", customvalidator.SetDefaultName)
	// validator.RegisterValidation("idn-mobile-number", customvalidator.SetIDNMobileNumber)
	// validator.RegisterValidation("ISO8601date", customvalidator.SetISO8601dateFormat)

	taskRepositoryV1 := taskV1.NewTaskRepository(logger, dbReadOnly, dbReadWrite, "task")
	taskUsecaseV1 := taskV1.NewTaskUsecase(logger, cfg.Application.Timezone, gcs, taskRepositoryV1)
	taskV1.NewTaskHTTPHandler(logger, router, basicAuthMiddleware, validator, taskUsecaseV1)

	taskRepositoryV2 := taskV2.NewTaskRepository(logger, dbReadOnly, dbReadWrite, "task")
	taskUsecaseV2 := taskV2.NewTaskUsecase(logger, cfg.Application.Timezone, gcs, taskRepositoryV2)
	taskV2.NewTaskHTTPHandler(logger, router, basicAuthMiddleware, validator, taskUsecaseV2)

	userRepository := user.NewUserRepository(logger, dbReadOnly, dbReadWrite, "user_encrypt")
	userUsecase := user.NewUserUsecase(logger, cfg.Application.Timezone, userRepository)
	user.NewUserHTTPHandler(logger, router, basicAuthMiddleware, validator, userUsecase)

	handler := middleware.ClientDeviceMiddleware(router)
	// set cors
	handler = cors.New(cors.Options{
		AllowedOrigins:   cfg.Application.AllowedOrigins,
		AllowedMethods:   []string{http.MethodPost, http.MethodGet, http.MethodPut, http.MethodDelete},
		AllowedHeaders:   []string{"Origin", "Accept", "Content-Type", "X-Requested-With", "Authorization"},
		AllowCredentials: true,
	}).Handler(handler)
	handler = middleware.NewRecovery(logger, true).Handler(handler)

	// initiate server
	srv := server.NewServer(logger, handler, cfg.Application.Port)
	srv.Start()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm

	srv.Close()
	dbReadOnly.Close()
	dbReadWrite.Close()

}

func index(w http.ResponseWriter, r *http.Request) {
	resp := response.NewSuccessResponse(nil, response.StatOK, indexMessage)
	response.JSON(w, resp)
}
