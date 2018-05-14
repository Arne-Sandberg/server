package grpcRouter

import (
	"context"
	"fmt"
	"os"
	"testing"
	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/db"
	"github.com/freecloudio/freecloud/fs"
	"github.com/freecloudio/freecloud/models"
	"google.golang.org/grpc"
	"gopkg.in/clog.v1"
)

func TestGrpcRouter(t *testing.T) {
	err := clog.New(clog.CONSOLE, clog.ConsoleConfig{
		Level: clog.TRACE,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not initialize logging: %v", err)
		os.Exit(2)
	}

	dfs, err := fs.NewDiskFilesystem("testData", "testTmp", 100)
	if err != nil {
		t.Errorf("Failed to initialize diskfilesystem: %v", err)
		return
	}

	database, err := db.NewStormDB("test.db")
	if err != nil {
		t.Errorf("Failed to initialize database: %v", err)
		return
	}

	auth.Init(database, database, 100)

	vfs, err := fs.NewVirtualFilesystem(dfs, database, "testTmp")
	if err != nil {
		t.Errorf("Failed to initialize virtualfilesystem: %v", err)
		return
	}

	Start(8081, 8082, "localhost", vfs, true)

	conn, err := grpc.Dial("localhost:8082", grpc.WithInsecure())
	if err != nil {
		t.Errorf("Failed to dial grpc server: %v", err)
		return
	}
	defer conn.Close()

	authClient := models.NewAuthServiceClient(conn)
	userClient := models.NewUserServiceClient(conn)
	systemClient := models.NewSystemServiceClient(conn)

	authResp, err := authClient.Signup(context.Background(), &models.User{FirstName: "Jon", LastName: "Doe", Email: "john.admin@testing.com", Password: "secretPassw0rd"})
	if err != nil {
		t.Errorf("Failed signup call: %v", err)
		return
	}

	authResp, err = authClient.Login(context.Background(), &models.User{Email: "john.admin@testing.com", Password: "secretPassw0rd"})
	if err != nil {
		t.Errorf("Failed login call: %v", err)
		return
	}
	adminToken := authResp.Token

	authResp, err = authClient.Signup(context.Background(), &models.User{FirstName: "Jon", LastName: "Doe", Email: "john.user@testing.com", Password: "secretPassw0rd"})
	if err != nil {
		t.Errorf("Failed signup call: %v", err)
		return
	}
	userToken := authResp.Token

	userResp, err := userClient.GetOwnUser(context.Background(), &models.Authentication{Token: adminToken})
	if err != nil {
		t.Errorf("Failed getting own user: %v", err)
	} else if userResp.IsAdmin == false {
		t.Error("First user is not an admin!")
	}

	userResp, err = userClient.GetUserByID(context.Background(), &models.UserIDRequest{Auth: &models.Authentication{Token: adminToken}, UserID: 2})
	if err != nil {
		t.Errorf("Failed getting user by ID: %v", err)
	} else if userResp.IsAdmin != false {
		t.Error("Second user is an admin!")
	}

	userResp, err = userClient.GetUserByEmail(context.Background(), &models.UserEmailRequest{Auth: &models.Authentication{Token: userToken}, UserEmail: "john.admin@testing.com"})
	if err != nil {
		t.Errorf("Failed getting user by email: %v", err)
	} else if userResp.Email != "john.admin@testing.com" {
		t.Errorf("Got email %v instead of email john.admin@testing.com", userResp.Email)
	}

	userResp, err = userClient.UpdateOwnUser(context.Background(), &models.UserUpdateRequest{Auth: &models.Authentication{Token: userToken}, UserUpdate: &models.UserUpdate{IsAdminOO: &models.UserUpdate_IsAdmin{IsAdmin: true}, FirstNameOO: &models.UserUpdate_FirstName{FirstName: "Peter"}}})
	if err != nil {
		t.Errorf("Failed updating user: %v", err)
	} else if userResp.IsAdmin != false {
		t.Errorf("Normal user could make itself an admin")
	} else if userResp.FirstName != "Peter" {
		t.Errorf("Changed firstName to %v instead of Peter", userResp.FirstName)
	}

	_, err = systemClient.GetSystemStats(context.Background(), &models.Authentication{Token: adminToken})
	if err != nil {
		t.Errorf("Failed to get systemStats: %v", err)
	}

	_, err = systemClient.GetSystemStats(context.Background(), &models.Authentication{Token: userToken})
	if err == nil {
		t.Error("User could get system stats")
	}

	_, err = userClient.DeleteOwnUser(context.Background(), &models.Authentication{Token: userToken})
	if err != nil {
		t.Errorf("Could not delete own user: %v", err)
	}

	_, err = authClient.Logout(context.Background(), &models.Authentication{Token: adminToken})
	if err != nil {
		t.Errorf("Failed logout call: %v", err)
		return
	}

	//Stop()
	vfs.Close()
	auth.Close()
	database.Close()
	dfs.Close()

	os.RemoveAll("testData")
	os.RemoveAll("testTmp")
	os.Remove("test.db")
	os.Remove("test.db.lock")
}
