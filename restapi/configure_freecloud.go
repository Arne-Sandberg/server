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

	"github.com/freecloudio/server/config"
	"github.com/freecloudio/server/manager"
	"github.com/freecloudio/server/restapi/operations"
	"github.com/freecloudio/server/restapi/operations/auth"
	"github.com/freecloudio/server/restapi/operations/file"
	"github.com/freecloudio/server/restapi/operations/system"
	"github.com/freecloudio/server/restapi/operations/user"
	"github.com/freecloudio/server/utils"

	"github.com/freecloudio/server/controller"
	"github.com/freecloudio/server/models"
	"github.com/freecloudio/server/repository"
)

const tmpName = ".tmp"

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
		//return controller.FileCreateHandler(params, principal)
		return middleware.NotImplemented("operation file.CreateFile has not yet been implemented")
	})
	api.UserDeleteCurrentUserHandler = user.DeleteCurrentUserHandlerFunc(func(params user.DeleteCurrentUserParams, principal *models.Principal) middleware.Responder {
		return controller.AuthDeleteCurrentUserHandler(params, principal)
	})
	api.FileDeleteFileHandler = file.DeleteFileHandlerFunc(func(params file.DeleteFileParams, principal *models.Principal) middleware.Responder {
		//return controller.FileDeleteHandler(params, principal)
		return middleware.NotImplemented("operation file.DeleteFile has not yet been implemented")
	})
	api.UserDeleteUserByUsernameHandler = user.DeleteUserByUsernameHandlerFunc(func(params user.DeleteUserByUsernameParams, principal *models.Principal) middleware.Responder {
		return controller.AuthDeleteUserByUsernameHandler(params, principal)
	})
	api.FileDownloadFileHandler = file.DownloadFileHandlerFunc(func(params file.DownloadFileParams, principal *models.Principal) middleware.Responder {
		return middleware.NotImplemented("operation file.DownloadFile has not yet been implemented")
	})
	api.UserGetCurrentUserHandler = user.GetCurrentUserHandlerFunc(func(params user.GetCurrentUserParams, principal *models.Principal) middleware.Responder {
		return controller.AuthGetCurrentUserHandler(params, principal)
	})
	api.FileGetPathInfoHandler = file.GetPathInfoHandlerFunc(func(params file.GetPathInfoParams, principal *models.Principal) middleware.Responder {
		//return controller.FileGetPathInfoHandler(params, principal)
		return middleware.NotImplemented("operation file.GetPathInfo has not yet been implemented")
	})
	api.SystemGetSystemStatsHandler = system.GetSystemStatsHandlerFunc(func(params system.GetSystemStatsParams, principal *models.Principal) middleware.Responder {
		return controller.SystemStatsHandler()
	})
	api.UserGetUserByUsernameHandler = user.GetUserByUsernameHandlerFunc(func(params user.GetUserByUsernameParams, principal *models.Principal) middleware.Responder {
		return controller.AuthGetUserByUsernameHandler(params, principal)
	})
	api.AuthLoginHandler = auth.LoginHandlerFunc(func(params auth.LoginParams) middleware.Responder {
		return controller.AuthLoginHandler(params)
	})
	api.AuthLogoutHandler = auth.LogoutHandlerFunc(func(params auth.LogoutParams, principal *models.Principal) middleware.Responder {
		return controller.AuthLogoutHandler(params, principal)
	})
	api.FileRescanCurrentUserHandler = file.RescanCurrentUserHandlerFunc(func(params file.RescanCurrentUserParams, principal *models.Principal) middleware.Responder {
		//return controller.FileRescanCurrentUserHandler(params, principal)
		return middleware.NotImplemented("operation file.RescanCurrentUser has not yet been implemented")
	})
	api.FileRescanUserByIDHandler = file.RescanUserByIDHandlerFunc(func(params file.RescanUserByIDParams, principal *models.Principal) middleware.Responder {
		//return controller.FileRescanUserByIDHandler(params, principal)
		return middleware.NotImplemented("operation file.RescanUserByID has not yet been implemented")
	})
	api.FileSearchFileHandler = file.SearchFileHandlerFunc(func(params file.SearchFileParams, principal *models.Principal) middleware.Responder {
		return middleware.NotImplemented("operation file.SearchFile has not yet been implemented")
	})
	api.FileShareFilesHandler = file.ShareFilesHandlerFunc(func(params file.ShareFilesParams, principal *models.Principal) middleware.Responder {
		//return controller.FileShareFilesHandler(params, principal)
		return middleware.NotImplemented("operation file.ShareFiles has not yet been implemented")
	})
	api.AuthSignupHandler = auth.SignupHandlerFunc(func(params auth.SignupParams) middleware.Responder {
		return controller.AuthSignupHandler(params)
	})
	api.FileGetStarredFileInfosHandler = file.GetStarredFileInfosHandlerFunc(func(params file.GetStarredFileInfosParams, principal *models.Principal) middleware.Responder {
		//return controller.FileGetStarredFileInfosHandler(params, principal)
		return middleware.NotImplemented("operation file.GetStarredFiles has not yet been implemented")
	})
	api.UserUpdateCurrentUserHandler = user.UpdateCurrentUserHandlerFunc(func(params user.UpdateCurrentUserParams, principal *models.Principal) middleware.Responder {
		return middleware.NotImplemented("operation user.UpdateCurrentUser has not yet been implemented")
	})
	api.FileUpdateFileHandler = file.UpdateFileHandlerFunc(func(params file.UpdateFileParams, principal *models.Principal) middleware.Responder {
		return middleware.NotImplemented("operation file.UpdateFile has not yet been implemented")
	})
	api.UserUpdateUserByUsernameHandler = user.UpdateUserByUsernameHandlerFunc(func(params user.UpdateUserByUsernameParams, principal *models.Principal) middleware.Responder {
		return middleware.NotImplemented("operation user.UpdateUserByUsername has not yet been implemented")
	})
	api.FileUploadFileHandler = file.UploadFileHandlerFunc(func(params file.UploadFileParams, principal *models.Principal) middleware.Responder {
		return middleware.NotImplemented("operation file.UploadFile has not yet been implemented")
	})
	api.FileZipFilesHandler = file.ZipFilesHandlerFunc(func(params file.ZipFilesParams, principal *models.Principal) middleware.Responder {
		//return controller.FileZipFilesHandler(params, principal)
		return middleware.NotImplemented("operation file.ZipFiles has not yet been implemented")
	})
	api.FileGetShareEntryByIDHandler = file.GetShareEntryByIDHandlerFunc(func(params file.GetShareEntryByIDParams, principal *models.Principal) middleware.Responder {
		//return controller.FileGetShareEntryByIDHandler(params, principal)
		return middleware.NotImplemented("operation file.GetShareEntryByID has not yet been implemented")
	})
	api.FileDeleteShareEntryByIDHandler = file.DeleteShareEntryByIDHandlerFunc(func(params file.DeleteShareEntryByIDParams, principal *models.Principal) middleware.Responder {
		//return controller.FileDeleteShareEntryByIDHandler(params, principal)
		return middleware.NotImplemented("operation file.GetShareEntryByID has not yet been implemented")
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
	err := repository.InitSQLDatabaseConnection(config.GetString("db.type"), config.GetString("db.user"), config.GetString("db.password"), config.GetString("db.host"), config.GetInt("db.port"), config.GetString("db.name"))
	if err != nil {
		log.Fatal(0, "Database setup failed, bailing out!: %v", err)
	}

	err = repository.InitGraphDatabaseConnection(config.GetString("graph.url"), config.GetString("graph.user"), config.GetString("graph.password"))
	if err != nil {
		log.Fatal(0, "Database setup failed, bailing out: %v", err)
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
	fileSystemRep, err := repository.CreateFileSystemRepository(config.GetString("fs.base_directory"), tmpName, config.GetInt("fs.tmp_clear_interval"), config.GetInt("fs.tmp_data_expiry"))
	if err != nil {
		log.Fatal(0, "FileSystemRepository setup failed, bailing out!: %v", err)
	}

	manager.CreateAuthManager(sessionRep, userRep, config.GetInt("auth.session_expiry"), config.GetInt("auth.session_cleanup_interval"))
	manager.CreateFileManager(fileSystemRep, fileInfoRep, shareEntryRep, tmpName)
	manager.CreateSystemManager("0.0.1") // TODO: Better place to save version
}

func shutdownServer() {
	manager.GetAuthManager().Close()
	repository.CloseSQLDatabaseConnection()
	repository.CloseSQLDatabaseConnection()
	utils.CloseLogger()
}
