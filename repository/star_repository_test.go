package repository

import (
	"os"
	"testing"

	"github.com/freecloudio/server/models"
)

func TestStarRepository(t *testing.T) {
	star0 := &models.Star{FileID: 1, UserID: 1}
	star1 := &models.Star{FileID: 1, UserID: 2}
	star2 := &models.Star{FileID: 2, UserID: 2}
	dbName := "starTest.db"

	cleanDBFiles := func() {
		os.Remove(dbName)
	}

	cleanDBFiles()
	defer cleanDBFiles()

	var rep *StarRepository

	success := t.Run("create connection and repository", func(t *testing.T) {
		err := InitDatabaseConnection("", "", "", "", 0, dbName)
		if err != nil {
			t.Fatalf("Failed to connect to gorm database: %v", err)
		}

		rep, err = CreateStarRepository()
		if err != nil {
			t.Fatalf("Failed to create user repository: %v", err)
		}
	})
	if !success {
		t.Skip("Further test skipped due to setup failing")
	}

	t.Run("empty repository", func(t *testing.T) {
		count, err := rep.Count()
		if err != nil {
			t.Fatalf("Failed to get count: %v", err)
		}
		if count > 0 {
			t.Errorf("Star count greater than zero for empty star repository: %d", count)
		}
	})

	success = t.Run("create stars", func(t *testing.T) {
		err := rep.Create(star0)
		if err != nil {
			t.Errorf("Failed to create star0: %v", err)
		}
		err = rep.Create(star1)
		if err != nil {
			t.Errorf("Failed to create star1: %v", err)
		}
		err = rep.Create(star2)
		if err != nil {
			t.Errorf("Failed to create star2: %v", err)
		}
	})
	if !success {
		t.Skip("Skipping further tests due to no created stars")
	}

	t.Run("correct count after creating stars", func(t *testing.T) {
		count, err := rep.Count()
		if err != nil {
			t.Fatalf("Failed to get count: %v", err)
		}
		if count != 3 {
			t.Errorf("Total count unequal to 3 for filled star repository: %d", err)
		}
	})

	t.Run("correct existenz of created stars", func(t *testing.T) {
		starExists, err := rep.Exists(star0.FileID, star0.UserID)
		if err != nil {
			t.Errorf("Failed to read whether star0 exists: %v", err)
		}
		if !starExists {
			t.Error("star0 does not exist but should have been created")
		}
		starExists, err = rep.Exists(star1.FileID, star1.UserID)
		if err != nil {
			t.Errorf("Failed to read whether star1 exists: %v", err)
		}
		if !starExists {
			t.Error("star1 does not exist but should have been created")
		}
		starExists, err = rep.Exists(star2.FileID, star2.UserID)
		if err != nil {
			t.Errorf("Failed to read whether star2 exists: %v", err)
		}
		if !starExists {
			t.Error("star2 does not exist but should have been created")
		}
	})

	delSuccess := t.Run("delete star", func(t *testing.T) {
		err := rep.Delete(star0.FileID, star0.UserID)
		if err != nil {
			t.Errorf("Failed to delete star0: %v", err)
		}
	})

	if delSuccess {
		t.Run("correct existenz after deleting star", func(t *testing.T) {
			starExists, err := rep.Exists(star0.FileID, star0.UserID)
			if err != nil {
				t.Errorf("Failed to read whether star0 exists: %v", err)
			}
			if starExists {
				t.Error("star0 does exists but should have been deleted")
			}
			starExists, err = rep.Exists(star1.FileID, star1.UserID)
			if err != nil {
				t.Errorf("Failed to read whether star1 exists: %v", err)
			}
			if !starExists {
				t.Error("star1 does not exist but should have been created and not deleted")
			}
			starExists, err = rep.Exists(star2.FileID, star2.UserID)
			if err != nil {
				t.Errorf("Failed to read whether star2 exists: %v", err)
			}
			if !starExists {
				t.Error("star2 does not exist but should have been created and not deleted")
			}
		})

		t.Run("correct count after user deletion", func(t *testing.T) {
			count, err := rep.Count()
			if err != nil {
				t.Fatalf("Failed to count stars after deletion: %v", err)
			}
			if count != 2 {
				t.Errorf("Count unequal to two after deletion: %d", count)
			}
		})
	}
}
