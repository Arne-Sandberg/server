// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"io"
	"net/http"

	errors "github.com/go-openapi/errors"
	runtime "github.com/go-openapi/runtime"
	middleware "github.com/go-openapi/runtime/middleware"
	clog "gopkg.in/clog.v1"

	"github.com/freecloudio/freecloud/packageInit"
	"github.com/freecloudio/freecloud/restapi/operations"
	"github.com/freecloudio/freecloud/restapi/operations/auth"
	"github.com/freecloudio/freecloud/restapi/operations/file"
	"github.com/freecloudio/freecloud/restapi/operations/system"
	"github.com/freecloudio/freecloud/restapi/operations/user"

	"github.com/freecloudio/freecloud/controller"
	"github.com/freecloudio/freecloud/models"
)

//go:generate swagger generate server --target .. --name Freecloud --spec ../api/freecloud.yml --principal models.User

func configureFlags(api *operations.FreecloudAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.FreecloudAPI) http.Handler {
	api.ServeError = errors.ServeError

	api.Logger = clog.Trace

	api.JSONConsumer = runtime.JSONConsumer()

	api.MultipartformConsumer = runtime.DiscardConsumer

	api.GzipProducer = runtime.ProducerFunc(func(w io.Writer, data interface{}) error {
		return errors.NotImplemented("gzip producer has not yet been implemented")
	})
	api.JSONProducer = runtime.JSONProducer()

	// Applies when the "Authorization" header is set
	api.TokenAuthAuth = func(token string, scopes []string) (*models.User, error) {
		return controller.ValidateSession(token, scopes)
	}

	api.FileCreateFileHandler = file.CreateFileHandlerFunc(func(params file.CreateFileParams, principal *models.User) middleware.Responder {
		return middleware.NotImplemented("operation file.CreateFile has not yet been implemented")
	})
	api.UserDeleteCurrentUserHandler = user.DeleteCurrentUserHandlerFunc(func(params user.DeleteCurrentUserParams, principal *models.User) middleware.Responder {
		return middleware.NotImplemented("operation user.DeleteCurrentUser has not yet been implemented")
	})
	api.FileDeleteFileHandler = file.DeleteFileHandlerFunc(func(params file.DeleteFileParams, principal *models.User) middleware.Responder {
		return middleware.NotImplemented("operation file.DeleteFile has not yet been implemented")
	})
	api.UserDeleteUserByIDHandler = user.DeleteUserByIDHandlerFunc(func(params user.DeleteUserByIDParams, principal *models.User) middleware.Responder {
		return middleware.NotImplemented("operation user.DeleteUserByID has not yet been implemented")
	})
	api.FileDownloadFileHandler = file.DownloadFileHandlerFunc(func(params file.DownloadFileParams, principal *models.User) middleware.Responder {
		return middleware.NotImplemented("operation file.DownloadFile has not yet been implemented")
	})
	api.UserGetCurrentUserHandler = user.GetCurrentUserHandlerFunc(func(params user.GetCurrentUserParams, principal *models.User) middleware.Responder {
		return middleware.NotImplemented("operation user.GetCurrentUser has not yet been implemented")
	})
	api.FileGetFileInfoHandler = file.GetFileInfoHandlerFunc(func(params file.GetFileInfoParams, principal *models.User) middleware.Responder {
		return middleware.NotImplemented("operation file.GetFileInfo has not yet been implemented")
	})
	api.SystemGetSystemStatsHandler = system.GetSystemStatsHandlerFunc(func(params system.GetSystemStatsParams, principal *models.User) middleware.Responder {
		return middleware.NotImplemented("operation system.GetSystemStats has not yet been implemented")
	})
	api.UserGetUserByIDHandler = user.GetUserByIDHandlerFunc(func(params user.GetUserByIDParams, principal *models.User) middleware.Responder {
		return middleware.NotImplemented("operation user.GetUserByID has not yet been implemented")
	})
	api.AuthLoginHandler = auth.LoginHandlerFunc(func(params auth.LoginParams) middleware.Responder {
		return middleware.NotImplemented("operation auth.Login has not yet been implemented")
	})
	api.AuthLogoutHandler = auth.LogoutHandlerFunc(func(params auth.LogoutParams, principal *models.User) middleware.Responder {
		return middleware.NotImplemented("operation auth.Logout has not yet been implemented")
	})
	api.FileRescanCurrentUserHandler = file.RescanCurrentUserHandlerFunc(func(params file.RescanCurrentUserParams, principal *models.User) middleware.Responder {
		return middleware.NotImplemented("operation file.RescanCurrentUser has not yet been implemented")
	})
	api.FileRescanUserByIDHandler = file.RescanUserByIDHandlerFunc(func(params file.RescanUserByIDParams, principal *models.User) middleware.Responder {
		return middleware.NotImplemented("operation file.RescanUserByID has not yet been implemented")
	})
	api.FileSearchFileHandler = file.SearchFileHandlerFunc(func(params file.SearchFileParams, principal *models.User) middleware.Responder {
		return middleware.NotImplemented("operation file.SearchFile has not yet been implemented")
	})
	api.FileShareFileHandler = file.ShareFileHandlerFunc(func(params file.ShareFileParams, principal *models.User) middleware.Responder {
		return middleware.NotImplemented("operation file.ShareFile has not yet been implemented")
	})
	api.AuthSignupHandler = auth.SignupHandlerFunc(func(params auth.SignupParams) middleware.Responder {
		return controller.AuthSignupHandler(params.Body)
	})
	api.FileStarredFilesHandler = file.StarredFilesHandlerFunc(func(params file.StarredFilesParams, principal *models.User) middleware.Responder {
		return middleware.NotImplemented("operation file.StarredFiles has not yet been implemented")
	})
	api.UserUpdateCurrentUserHandler = user.UpdateCurrentUserHandlerFunc(func(params user.UpdateCurrentUserParams, principal *models.User) middleware.Responder {
		return middleware.NotImplemented("operation user.UpdateCurrentUser has not yet been implemented")
	})
	api.FileUpdateFileHandler = file.UpdateFileHandlerFunc(func(params file.UpdateFileParams, principal *models.User) middleware.Responder {
		return middleware.NotImplemented("operation file.UpdateFile has not yet been implemented")
	})
	api.UserUpdateUserByIDHandler = user.UpdateUserByIDHandlerFunc(func(params user.UpdateUserByIDParams, principal *models.User) middleware.Responder {
		return middleware.NotImplemented("operation user.UpdateUserByID has not yet been implemented")
	})
	api.FileUploadFileHandler = file.UploadFileHandlerFunc(func(params file.UploadFileParams, principal *models.User) middleware.Responder {
		return middleware.NotImplemented("operation file.UploadFile has not yet been implemented")
	})
	api.FileZipFilesHandler = file.ZipFilesHandlerFunc(func(params file.ZipFilesParams, principal *models.User) middleware.Responder {
		return middleware.NotImplemented("operation file.ZipFiles has not yet been implemented")
	})

	packageInit.Init()

	api.ServerShutdown = func() {
		packageInit.Deinit()
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
	return controller.FileServerMiddleware(handler)
}
