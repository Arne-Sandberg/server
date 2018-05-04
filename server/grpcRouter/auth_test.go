package grpcRouter

import (
	"testing"
	"context"
	"google.golang.org/grpc"
	"github.com/freecloudio/freecloud/models"
	"math/rand"
	"strconv"
	"net/http"
	"github.com/freecloudio/freecloud/fs"
	"github.com/freecloudio/freecloud/db"
	"github.com/freecloudio/freecloud/auth"
	"time"
	"os"
)

func TestAuthService(t *testing.T) {
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

	Start(8081, "localhost", vfs)

	conn, err := grpc.Dial("localhost:8081", grpc.WithInsecure())
	if err != nil {
		t.Errorf("Failed to dial grpc server: %v", err)
		return
	}
	defer conn.Close()

	authClient := models.NewAuthServiceClient(conn)

	rand.Seed(time.Now().Unix())
	email := "john.doe" + strconv.Itoa(rand.Int()) + "@testing.com"

	authResp, err := authClient.Signup(context.Background(), &models.User{ FirstName: "Jon", LastName: "Doe", Email: email, Password: "secretPassw0rd" })
	if err != nil {
		t.Errorf("Failed signup call: %v", err)
		return
	}

	if authResp.Meta.ResponseCode != http.StatusCreated || authResp.Token == "" {
		t.Errorf("Signup response not correct: Got %d instead of %d: %s", authResp.Meta.ResponseCode, http.StatusCreated, authResp.Meta.ErrorMessage)
		return
	}

	authResp, err = authClient.Login(context.Background(), &models.User{ Email: email, Password: "secretPassw0rd" })
	if err != nil {
		t.Errorf("Failed login call: %v", err)
		return
	}

	if authResp.Meta.ResponseCode != http.StatusOK || authResp.Token == "" {
		t.Errorf("Login response not correct: Got %d instead of %d: %s", authResp.Meta.ResponseCode, http.StatusOK, authResp.Meta.ErrorMessage)
		return
	}

	resp, err := authClient.Logout(context.Background(), &models.Authentication{ Token: authResp.Token })
	if err != nil {
		t.Errorf("Failed logout call: %v", err)
		return
	}

	if resp.ResponseCode != http.StatusOK {
		t.Errorf("Logout response not correct: Got %d instead of %d: %s", resp.ResponseCode, http.StatusOK, resp.ErrorMessage)
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
