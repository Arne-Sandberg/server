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

var vfs *fs.VirtualFilesystem
var dfs *fs.DiskFilesystem
var database *db.XormDB

func SetupTest() error {
	err := clog.New(clog.CONSOLE, clog.ConsoleConfig{
		Level: clog.TRACE,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not initialize logging: %v", err)
		os.Exit(2)
	}

	dfs, err = fs.NewDiskFilesystem("testData", 100)
	if err != nil {
		return fmt.Errorf("failed to initialize diskfilesystem: %v", err)
	}

	database, err = db.NewXormDB("test.db")
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}

	auth.Init(database, database, 100)

	vfs, err = fs.NewVirtualFilesystem(dfs, database)
	if err != nil {
		return fmt.Errorf("failed to initialize virtualfilesystem: %v", err)
	}

	Start(8081, 8082, "localhost", vfs, true)

	return nil
}

func FinishTest() {
	//Stop()
	vfs.Close()
	auth.Close()
	database.Close()
	dfs.Close()
}

func DeleteTestFiles() {
	os.RemoveAll("testData")
	os.RemoveAll("testTmp")
	os.Remove("test.db")
	os.Remove("test.db.lock")
}

func TestGrpcRouter(t *testing.T) {
	DeleteTestFiles()

	err := SetupTest()
	if err != nil {
		t.Errorf("Setup failed: %v", err)
		return
	}

	conn, err := grpc.Dial("localhost:8082", grpc.WithInsecure())
	if err != nil {
		t.Errorf("Failed to dial grpc server: %v", err)
		return
	}
	defer conn.Close()

	authClient := models.NewAuthServiceClient(conn)
	userClient := models.NewUserServiceClient(conn)
	filesClient := models.NewFilesServiceClient(conn)
	systemClient := models.NewSystemServiceClient(conn)

	// Auth
	authResp, err := authClient.Signup(context.Background(), &models.User{FirstName: "Jon", LastName: "Doe", Email: "john.admin@testing.com", Password: "secretPassw0rd"})
	if err != nil {
		t.Errorf("Failed signup call: %v", err)
		return
	}

	authResp, err = authClient.Login(context.Background(), &models.LoginData{Email: "john.admin@testing.com", Password: "secretPassw0rd"})
	if err != nil {
		t.Errorf("Failed login call: %v", err)
		return
	}
	adminAuth := &models.Authentication{Token: authResp.Token}

	authResp, err = authClient.Signup(context.Background(), &models.User{FirstName: "Jon", LastName: "Doe", Email: "john.user@testing.com", Password: "secretPassw0rd"})
	if err != nil {
		t.Errorf("Failed signup call: %v", err)
		return
	}
	userAuth := &models.Authentication{Token: authResp.Token}

	// User
	userResp, err := userClient.GetOwnUser(context.Background(), adminAuth)
	if err != nil {
		t.Errorf("Failed getting own user: %v", err)
	} else if userResp.IsAdmin == false {
		t.Error("First user is not an admin!")
	}

	userResp, err = userClient.GetUserByID(context.Background(), &models.UserIDRequest{Auth: adminAuth, UserID: 2})
	if err != nil {
		t.Errorf("Failed getting user by ID: %v", err)
	} else if userResp.IsAdmin != false {
		t.Error("Second user is an admin!")
	}

	userResp, err = userClient.GetUserByEmail(context.Background(), &models.UserEmailRequest{Auth: userAuth, UserEmail: "john.admin@testing.com"})
	if err != nil {
		t.Errorf("Failed getting user by email: %v", err)
	} else if userResp.Email != "john.admin@testing.com" {
		t.Errorf("Got email %v instead of email john.admin@testing.com", userResp.Email)
	}

	userResp, err = userClient.UpdateOwnUser(context.Background(), &models.UserUpdateRequest{Auth: userAuth, UserUpdate: &models.UserUpdate{IsAdminOO: &models.UserUpdate_IsAdmin{IsAdmin: true}, FirstNameOO: &models.UserUpdate_FirstName{FirstName: "Peter"}}})
	if err != nil {
		t.Errorf("Failed updating user: %v", err)
	} else if userResp.IsAdmin != false {
		t.Errorf("Normal user could make itself an admin")
	} else if userResp.FirstName != "Peter" {
		t.Errorf("Changed firstName to %v instead of Peter", userResp.FirstName)
	}

	// Files
	_, err = filesClient.CreateFile(context.Background(), &models.CreateFileRequest{Auth: adminAuth, IsDir: false, FullPath: "/testFile.txt"})
	if err != nil {
		t.Errorf("Failed to create file: %v", err)
	}

	_, err = filesClient.CreateFile(context.Background(), &models.CreateFileRequest{Auth: adminAuth, IsDir: false, FullPath: "/testFile.txt"})
	if err == nil {
		t.Errorf("Could create duplicate file: %v", err)
	}

	_, err = filesClient.ShareFiles(context.Background(), &models.ShareRequest{Auth: adminAuth, FullPaths: []string{"/testFile.txt"}, UserIDs: []uint32{2}})
	if err != nil {
		t.Errorf("Could not share file with other user: %v", err)
	}

	_, err = filesClient.GetFileInfo(context.Background(), &models.PathRequest{Auth: userAuth, FullPath: "/testFile.txt"})
	if err != nil {
		t.Errorf("Could not get shared file info: %v", err)
	}

	_, err = filesClient.CreateFile(context.Background(), &models.CreateFileRequest{Auth: adminAuth, IsDir: true, FullPath: "/testDir"})
	if err != nil {
		t.Errorf("Failed to create folder: %v", err)
	}

	_, err = filesClient.CreateFile(context.Background(), &models.CreateFileRequest{Auth: adminAuth, IsDir: false, FullPath: "/testDir/testFile.txt"})
	if err != nil {
		t.Errorf("Failed to create file in folder: %v", err)
	}

	_, err = filesClient.ShareFiles(context.Background(), &models.ShareRequest{Auth: adminAuth, FullPaths: []string{"/testDir"}, UserIDs: []uint32{2}})
	if err != nil {
		t.Errorf("Could not share folder with other user: %v", err)
	}

	_, err = filesClient.GetFileInfo(context.Background(), &models.PathRequest{Auth: userAuth, FullPath: "/testDir"})
	if err != nil {
		t.Errorf("Could not get shared folder info: %v", err)
	}

	_, err = filesClient.GetFileInfo(context.Background(), &models.PathRequest{Auth: userAuth, FullPath: "/testDir/testFile.txt"})
	if err != nil {
		t.Errorf("Could not get file info inside shared folder: %v", err)
	}

	_, err = filesClient.CreateFile(context.Background(), &models.CreateFileRequest{Auth: adminAuth, IsDir: true, FullPath: "/testDir/subDir"})
	if err != nil {
		t.Errorf("Failed to create subfolder in folder: %v", err)
	}

	_, err = filesClient.CreateFile(context.Background(), &models.CreateFileRequest{Auth: adminAuth, IsDir: false, FullPath: "/testDir/subDir/subDirFile.txt"})
	if err != nil {
		t.Errorf("Failed to create file in subfolder: %v", err)
	}

	_, err = filesClient.GetFileInfo(context.Background(), &models.PathRequest{Auth: userAuth, FullPath: "/testDir/subDir/subDirFile.txt"})
	if err != nil {
		t.Errorf("Could not get file info inside shared subfolder: %v", err)
	}

	// System stats
	_, err = systemClient.GetSystemStats(context.Background(), adminAuth)
	if err != nil {
		t.Errorf("Failed to get systemStats: %v", err)
	}

	_, err = systemClient.GetSystemStats(context.Background(), userAuth)
	if err == nil {
		t.Error("User could get system stats")
	}

	// User and Auth cleanup
	_, err = userClient.DeleteOwnUser(context.Background(), userAuth)
	if err != nil {
		t.Errorf("Could not delete own user: %v", err)
	}

	_, err = authClient.Logout(context.Background(), adminAuth)
	if err != nil {
		t.Errorf("Failed logout call: %v", err)
		return
	}

	FinishTest()
	DeleteTestFiles()
}
