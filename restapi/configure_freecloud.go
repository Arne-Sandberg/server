// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"io"
	"net/http"

	errors "github.com/go-openapi/errors"
	runtime "github.com/go-openapi/runtime"
	middleware "github.com/go-openapi/runtime/middleware"
	log "gopkg.in/clog.v1"

	"github.com/freecloudio/freecloud/config"
	"github.com/freecloudio/freecloud/manager"
	"github.com/freecloudio/freecloud/restapi/operations"
	"github.com/freecloudio/freecloud/restapi/operations/auth"
	"github.com/freecloudio/freecloud/restapi/operations/file"
	"github.com/freecloudio/freecloud/restapi/operations/system"
	"github.com/freecloudio/freecloud/restapi/operations/user"
	"github.com/freecloudio/freecloud/utils"

	"github.com/freecloudio/freecloud/controller"
	"github.com/freecloudio/freecloud/models"
	"github.com/freecloudio/freecloud/repository"
)

//go:generate swagger generate server --name Freecloud --spec ./api/freecloud.yml --principal models.Principal

func configureFlags(api *operations.FreecloudAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.FreecloudAPI) http.Handler {
	utils.SetupLogger()

	api.ServeError = errors.ServeError

	api.Logger = log.Trace

	api.JSONConsumer = runtime.JSONConsumer()

	api.MultipartformConsumer = MultipartformConsumer()

	api.GzipProducer = runtime.ProducerFunc(func(w io.Writer, data interface{}) error {
		return errors.NotImplemented("gzip producer has not yet been implemented")
	})
	api.JSONProducer = runtime.JSONProducer()

	// Applies when the "Authorization" header is set
	api.TokenAuthAuth = func(token string, scopes []string) (*models.Principal, error) {
		return controller.ValidateToken(token, scopes)
	}

	api.FileCreateFileHandler = file.CreateFileHandlerFunc(func(params file.CreateFileParams, principal *models.Principal) middleware.Responder {
		return controller.FileCreateHandler(params, principal)
	})
	api.UserDeleteCurrentUserHandler = user.DeleteCurrentUserHandlerFunc(func(params user.DeleteCurrentUserParams, principal *models.Principal) middleware.Responder {
		return controller.AuthDeleteCurrentUserHandler(params, principal)
	})
	api.FileDeleteFileHandler = file.DeleteFileHandlerFunc(func(params file.DeleteFileParams, principal *models.Principal) middleware.Responder {
		return controller.FileDeleteHandler(params, principal)
	})
	api.UserDeleteUserByIDHandler = user.DeleteUserByIDHandlerFunc(func(params user.DeleteUserByIDParams, principal *models.Principal) middleware.Responder {
		return controller.AuthDeleteUserByIDHandler(params, principal)
	})
	api.FileDownloadFileHandler = file.DownloadFileHandlerFunc(func(params file.DownloadFileParams, principal *models.Principal) middleware.Responder {
		return middleware.NotImplemented("operation file.DownloadFile has not yet been implemented")
	})
	api.UserGetCurrentUserHandler = user.GetCurrentUserHandlerFunc(func(params user.GetCurrentUserParams, principal *models.Principal) middleware.Responder {
		return controller.AuthGetCurrentUserHandler(params, principal)
	})
	api.FileGetPathInfoHandler = file.GetPathInfoHandlerFunc(func(params file.GetPathInfoParams, principal *models.Principal) middleware.Responder {
		return controller.FileGetPathInfoHandler(params, principal)
	})
	api.SystemGetSystemStatsHandler = system.GetSystemStatsHandlerFunc(func(params system.GetSystemStatsParams, principal *models.Principal) middleware.Responder {
		return controller.SystemStatsHandler()
	})
	api.UserGetUserByIDHandler = user.GetUserByIDHandlerFunc(func(params user.GetUserByIDParams, principal *models.Principal) middleware.Responder {
		return controller.AuthGetUserByIDHandler(params, principal)
	})
	api.AuthLoginHandler = auth.LoginHandlerFunc(func(params auth.LoginParams) middleware.Responder {
		return controller.AuthLoginHandler(params)
	})
	api.AuthLogoutHandler = auth.LogoutHandlerFunc(func(params auth.LogoutParams, principal *models.Principal) middleware.Responder {
		return controller.AuthLogoutHandler(params, principal)
	})
	api.FileRescanCurrentUserHandler = file.RescanCurrentUserHandlerFunc(func(params file.RescanCurrentUserParams, principal *models.Principal) middleware.Responder {
		return controller.FileRescanCurrentUserHandler(params, principal)
	})
	api.FileRescanUserByIDHandler = file.RescanUserByIDHandlerFunc(func(params file.RescanUserByIDParams, principal *models.Principal) middleware.Responder {
		return controller.FileRescanUserByIDHandler(params, principal)
	})
	api.FileSearchFileHandler = file.SearchFileHandlerFunc(func(params file.SearchFileParams, principal *models.Principal) middleware.Responder {
		return middleware.NotImplemented("operation file.SearchFile has not yet been implemented")
	})
	api.FileShareFileHandler = file.ShareFileHandlerFunc(func(params file.ShareFileParams, principal *models.Principal) middleware.Responder {
		return middleware.NotImplemented("operation file.ShareFile has not yet been implemented")
	})
	api.AuthSignupHandler = auth.SignupHandlerFunc(func(params auth.SignupParams) middleware.Responder {
		return controller.AuthSignupHandler(params)
	})
	api.FileGetStarredFileInfosHandler = file.GetStarredFileInfosHandlerFunc(func(params file.GetStarredFileInfosParams, principal *models.Principal) middleware.Responder {
		return controller.FileGetStarredFileInfosHandler(params, principal)
	})
	api.UserUpdateCurrentUserHandler = user.UpdateCurrentUserHandlerFunc(func(params user.UpdateCurrentUserParams, principal *models.Principal) middleware.Responder {
		return middleware.NotImplemented("operation user.UpdateCurrentUser has not yet been implemented")
	})
	api.FileUpdateFileHandler = file.UpdateFileHandlerFunc(func(params file.UpdateFileParams, principal *models.Principal) middleware.Responder {
		return middleware.NotImplemented("operation file.UpdateFile has not yet been implemented")
	})
	api.UserUpdateUserByIDHandler = user.UpdateUserByIDHandlerFunc(func(params user.UpdateUserByIDParams, principal *models.Principal) middleware.Responder {
		return middleware.NotImplemented("operation user.UpdateUserByID has not yet been implemented")
	})
	api.FileUploadFileHandler = file.UploadFileHandlerFunc(func(params file.UploadFileParams, principal *models.Principal) middleware.Responder {
		return middleware.NotImplemented("operation file.UploadFile has not yet been implemented")
	})
	api.FileZipFilesHandler = file.ZipFilesHandlerFunc(func(params file.ZipFilesParams, principal *models.Principal) middleware.Responder {
		return controller.FileZipFilesHandler(params, principal)
	})

	initializeServer()
	api.ServerShutdown = func() {
		shutdownServer()
	}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *http.Server, scheme, addr string) {

}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	fileServer := controller.FileServerMiddleware(handler)
	return controller.LoggingMiddleware(fileServer)
}

func initializeServer() {
	config.Init()

	err := repository.InitDatabaseConnection(config.GetString("db.type"), config.GetString("db.user"), config.GetString("db.password"), config.GetString("db.host"), config.GetInt("db.port"), config.GetString("db.name"))
	if err != nil {
		log.Fatal(0, "Database setup failed, bailing out!: %v", err)
	}

	userRep, err := repository.CreateUserRepository()
	if err != nil {
		log.Fatal(0, "UserRepository setup failed, bailing out!: %v", err)
	}
	sessionRep, err := repository.CreateSessionRepository()
	if err != nil {
		log.Fatal(0, "SessionRepository setup failed, bailing out!: %v", err)
	}
	fileInfoRep, err := repository.CreateFileInfoRepository()
	if err != nil {
		log.Fatal(0, "FileInfoRepository setup failed, bailing out!: %v", err)
	}
	shareEntryRep, err := repository.CreateShareEntryRepository()
	if err != nil {
		log.Fatal(0, "ShareEntryRepository setup failed, bailing out!: %v", err)
	}
	fileSystemRep, err := repository.CreateFileSystemRepository(config.GetString("fs.base_directory"), config.GetInt("fs.tmp_data_expiry"))
	if err != nil {
		log.Fatal(0, "FileSystemRepository setup failed, bailing out!: %v", err)
	}

	manager.CreateAuthManager(sessionRep, userRep)
	manager.CreateFileManager(fileSystemRep, fileInfoRep, shareEntryRep)
	manager.CreateStatsManager("0.0.1") // TODO: Better place to save version
}

func shutdownServer() {
	manager.GetAuthManager().Close()
	repository.CloseDatabaseConnection()
	utils.CloseLogger()
}
