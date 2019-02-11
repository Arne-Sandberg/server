package repository

import (
	"os"
	"reflect"
	"testing"

	"github.com/freecloudio/server/models"
)

var testUserSetupFailed = false
var testUserDBName = "userTest.db"
var testUserAdmin = &models.User{Email: "admin.user@example.com", IsAdmin: true}
var testUser0 = &models.User{Email: "user1@example.com"}
var testUser1 = &models.User{Email: "user2@example.com"}

func testUserCleanup() {
	os.Remove(testUserDBName)
}

func testUserSetup() *UserRepository {
	testUserCleanup()
	InitDatabaseConnection("", "", "", "", 0, testUserDBName)
	rep, _ := CreateUserRepository()
	return rep
}

func testUserInsert(rep *UserRepository) {
	rep.Create(testUserAdmin)
	rep.Create(testUser0)
	rep.Create(testUser1)
}

func TestCreateUserRepository(t *testing.T) {
	testUserCleanup()
	defer testUserCleanup()

	err := InitDatabaseConnection("", "", "", "", 0, testUserDBName)
	if err != nil {
		t.Errorf("Failed to connect to gorm database: %v", err)
	}

	_, err = CreateUserRepository()
	if err != nil {
		t.Errorf("Failed to create user repository: %v", err)
	}

	if t.Failed() {
		testUserSetupFailed = true
	}

}

func TestCreateUser(t *testing.T) {
	if testUserSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	defer testUserCleanup()
	rep := testUserSetup()

	err := rep.Create(testUserAdmin)
	if err != nil {
		t.Errorf("Failed to create admin user: %v", err)
	}

	if t.Failed() {
		testUserSetupFailed = true
	}
}

func TestCountUsers(t *testing.T) {
	if testUserSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	defer testUserCleanup()
	rep := testUserSetup()

	adminCount, err := rep.AdminCount()
	if err != nil {
		t.Errorf("Failed to get admin count: %v", err)
	}
	if adminCount > 0 {
		t.Errorf("Admin count greater than zero for empty user repository: %d", adminCount)
	}
	totalCount, err := rep.TotalCount()
	if err != nil {
		t.Errorf("Failed to get total count: %v", err)
	}
	if totalCount > 0 {
		t.Errorf("Total count greater than zero for empty user repository: %d", totalCount)
	}

	testUserInsert(rep)

	adminCount, err = rep.AdminCount()
	if err != nil {
		t.Errorf("Failed to get admin count: %v", err)
	}
	if adminCount != 1 {
		t.Errorf("Admin count unequal to 1 for filled user repository: %d", adminCount)
	}
	totalCount, err = rep.TotalCount()
	if err != nil {
		t.Errorf("Failed to get total count: %v", err)
	}
	if totalCount != 3 {
		t.Errorf("Total count unequal to 3 for filled user repository: %d", totalCount)
	}
}

func TestUserGetByID(t *testing.T) {
	if testUserSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	defer testUserCleanup()
	rep := testUserSetup()

	testUserInsert(rep)

	readBackUser, err := rep.GetByID(testUserAdmin.ID)
	if err != nil {
		t.Fatalf("Failed to read back admin user by ID: %v", err)
	}
	if !reflect.DeepEqual(readBackUser, testUserAdmin) {
		t.Errorf("Read back admin user and admin user not deeply equal: %v != %v", readBackUser, testUserAdmin)
	}
}

func TestUserGetByEmail(t *testing.T) {
	if testUserSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	defer testUserCleanup()
	rep := testUserSetup()

	testUserInsert(rep)

	readBackUser, err := rep.GetByEmail(testUser0.Email)
	if err != nil {
		t.Fatalf("Failed to read back user0 by Email: %v", err)
	}
	if !reflect.DeepEqual(readBackUser, testUser0) {
		t.Errorf("Read back user0 and user0 not deeply equal: %v != %v", readBackUser, testUser0)
	}
}

func TestGetAllUsers(t *testing.T) {
	if testUserSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	defer testUserCleanup()
	rep := testUserSetup()

	testUserInsert(rep)

	allUsers, err := rep.GetAll()
	if err != nil {
		t.Errorf("Failed to get all users: %v", err)
	}
	if len(allUsers) != 3 {
		t.Errorf("Length of read back users unequal to 3: %d", len(allUsers))
	}
}

func TestDeleteUser(t *testing.T) {
	if testUserSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	defer testUserCleanup()
	rep := testUserSetup()

	testUserInsert(rep)

	err := rep.Delete(testUser0.ID)
	if err != nil {
		t.Errorf("Failed to delete user1: %v", err)
	}

	_, err = rep.GetByID(testUser0.ID)
	if err == nil || !IsRecordNotFoundError(err) {
		t.Errorf("Succeeded to read deleted user by ID or error is not 'record not found': %v", err)
	}
	_, err = rep.GetByEmail(testUser0.Email)
	if err == nil || !IsRecordNotFoundError(err) {
		t.Errorf("Succeeded to read deleted user by Email or error is not 'record not found': %v", err)
	}
	allUsers, err := rep.GetAll()
	if err != nil {
		t.Errorf("Failed to get all users: %v", err)
	}
	if len(allUsers) != 2 {
		t.Errorf("Length of read back users unequal to 2 after deletion of user1: %d", len(allUsers))
	}

	adminCount, err := rep.AdminCount()
	if err != nil {
		t.Errorf("Failed to get admin count: %v", err)
	}
	if adminCount != 1 {
		t.Errorf("Admin count unequal to 1 for filled user repository: %d", adminCount)
	}

	totalCount, err := rep.TotalCount()
	if err != nil {
		t.Errorf("Failed to get total count: %v", err)
	}
	if totalCount != 2 {
		t.Errorf("Total count unequal to 2 for filled user repository: %d", totalCount)
	}
}

func TestUpdateUser(t *testing.T) {
	if testUserSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	defer testUserCleanup()
	rep := testUserSetup()

	testUserInsert(rep)

	testUser1.Email = "updatedEmail@example.com"
	err := rep.Update(testUser1)
	if err != nil {
		t.Errorf("Failed to update testUser1: %v", err)
	}

	readBackUser, err := rep.GetByID(testUser1.ID)
	if err != nil {
		t.Errorf("Failed to read back updated testUser1 by ID: %v", err)
	}
	if !reflect.DeepEqual(readBackUser, testUser1) {
		t.Error("Read back updated testUser1 by ID and testUser1 are not deeply equal")
	}
	readBackUser, err = rep.GetByEmail(testUser1.Email)
	if err != nil {
		t.Errorf("Failed to read back updated testUser1 by Email: %v", err)
	}
	if !reflect.DeepEqual(readBackUser, testUser1) {
		t.Error("Read back updated testUser1 by Email and testUser1 are not deeply equal")
	}
}
