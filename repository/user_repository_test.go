package repository

import (
	"os"
	"reflect"
	"testing"

	"github.com/freecloudio/server/models"
)

func cleanDBFiles() {
	os.Remove("userTest.db")
}

var adminUser = &models.User{Email: "admin.user@example.com", IsAdmin: true}
var user1 = &models.User{Email: "user1@example.com"}
var user2 = &models.User{Email: "user2@example.com"}

func TestUserRepository(t *testing.T) {
	cleanDBFiles()

	var rep *UserRepository

	success := t.Run("create connection and repository", func(t *testing.T) {
		err := InitDatabaseConnection("", "", "", "", 0, "userTest.db")
		if err != nil {
			t.Fatalf("Failed to connect to gorm database: %v", err)
		}

		rep, err = CreateUserRepository()
		if err != nil {
			t.Errorf("Failed to create user repository: %v", err)
		}
	})
	if !success {
		cleanDBFiles()
		t.Skip("Further test skipped due to setup failing")
	}

	t.Run("empty repository", func(t *testing.T) {
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
	})

	success = t.Run("create users", func(t *testing.T) {
		err := rep.Create(adminUser)
		if err != nil {
			t.Errorf("Failed to create admin user: %v", err)
		}
		err = rep.Create(user1)
		if err != nil {
			t.Errorf("Failed to create user1: %v", err)
		}
		err = rep.Create(user2)
		if err != nil {
			t.Errorf("Failed to create user2: %v", err)
		}
	})
	if !success {
		cleanDBFiles()
		t.Skip("Skipping further tests due to no created users")
	}

	t.Run("correct counts after creating users", func(t *testing.T) {
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
		if totalCount != 3 {
			t.Errorf("Total count unequal to 3 for filled user repository: %d", totalCount)
		}
	})

	t.Run("correct read back of created users", func(t *testing.T) {
		readBackUser, err := rep.GetByID(adminUser.ID)
		if err != nil {
			t.Errorf("Failed to read back admin user by ID: %v", err)
		}
		if !reflect.DeepEqual(readBackUser, adminUser) {
			t.Error("Read back admin user and admin user not deeply equal")
		}
		readBackUser, err = rep.GetByEmail(user1.Email)
		if err != nil {
			t.Errorf("Failed to read back user1 by Email: %v", err)
		}
		if !reflect.DeepEqual(readBackUser, user1) {
			t.Error("Read back user1 and user1 not deeply equal")
		}
		allUsers, err := rep.GetAll()
		if err != nil {
			t.Errorf("Failed to get all users: %v", err)
		}
		if len(allUsers) != 3 {
			t.Errorf("Length of read back users unequal to 3: %d", len(allUsers))
		}
	})

	delSuccess := t.Run("deleting user", func(t *testing.T) {
		err := rep.Delete(user1.ID)
		if err != nil {
			t.Errorf("Failed to delete user1: %v", err)
		}
	})

	if delSuccess {
		t.Run("correct read back of deleted user", func(t *testing.T) {
			_, err := rep.GetByID(user1.ID)
			if err == nil || !IsRecordNotFoundError(err) {
				t.Errorf("Succeeded to read deleted user by ID or error is not 'record not found': %v", err)
			}
			_, err = rep.GetByEmail(user1.Email)
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
		})
	}

	if delSuccess {
		t.Run("correct counts after user deletion", func(t *testing.T) {
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
		})
	}

	user2.Email = "updatedEmail@example.com"
	updSuccess := t.Run("updating user", func(t *testing.T) {
		err := rep.Update(user2)
		if err != nil {
			t.Errorf("Failed to update user2: %v", err)
		}
	})

	if updSuccess {
		t.Run("correct read back after updating user", func(t *testing.T) {
			readBackUser, err := rep.GetByID(user2.ID)
			if err != nil {
				t.Errorf("Failed to read back updated user2 by ID: %v", err)
			}
			if !reflect.DeepEqual(readBackUser, user2) {
				t.Error("Read back updated user2 by ID and user2 are not deeply equal")
			}
			readBackUser, err = rep.GetByEmail(user2.Email)
			if err != nil {
				t.Errorf("Failed to read back updated user2 by Email: %v", err)
			}
			if !reflect.DeepEqual(readBackUser, user2) {
				t.Error("Read back updated user2 by Email and user2 are not deeply equal")
			}
		})
	}

	cleanDBFiles()
}
