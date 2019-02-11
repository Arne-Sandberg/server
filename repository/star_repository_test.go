package repository

import (
	"os"
	"testing"

	"github.com/freecloudio/server/models"
)

var testStarSetupFailed = false
var testStarDBName = "starTest.db"
var testStar0 = &models.Star{FileID: 1, UserID: 1}
var testStar1 = &models.Star{FileID: 1, UserID: 2}

func testStarCleanup() {
	os.Remove(testStarDBName)
}

func testStarSetup() *StarRepository {
	testStarCleanup()
	InitDatabaseConnection("", "", "", "", 0, testStarDBName)
	rep, _ := CreateStarRepository()
	return rep
}

func testStarInsert(rep *StarRepository) {
	rep.Create(testStar0)
	rep.Create(testStar1)
}

func TestCreateStarRepository(t *testing.T) {
	testStarCleanup()
	defer testStarCleanup()

	err := InitDatabaseConnection("", "", "", "", 0, testStarDBName)
	if err != nil {
		t.Errorf("Failed to connect to gorm database: %v", err)
	}

	_, err = CreateStarRepository()
	if err != nil {
		t.Errorf("Failed to create star repository: %v", err)
	}

	if t.Failed() {
		testStarSetupFailed = true
	}
}

func TestCreateStars(t *testing.T) {
	if testStarSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testStarCleanup()
	rep := testStarSetup()

	err := rep.Create(testStar0)
	if err != nil {
		t.Errorf("Failed to create star0: %v", err)
	}

	if t.Failed() {
		testStarSetupFailed = true
	}
}

func TestCountStars(t *testing.T) {
	if testStarSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testStarCleanup()
	rep := testStarSetup()

	count, err := rep.Count()
	if err != nil {
		t.Fatalf("Failed to get count: %v", err)
	}
	if count > 0 {
		t.Errorf("Star count greater than zero for empty star repository: %d", count)
	}

	testStarInsert(rep)

	count, err = rep.Count()
	if err != nil {
		t.Fatalf("Failed to get count: %v", err)
	}
	if count != 2 {
		t.Errorf("Total count unequal to 2 for filled star repository: %d", err)
	}
}

func TestExistsStar(t *testing.T) {
	if testStarSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testStarCleanup()
	rep := testStarSetup()

	testStarInsert(rep)

	starExists, err := rep.Exists(testStar0.FileID, testStar0.UserID)
	if err != nil {
		t.Errorf("Failed to read whether star0 exists: %v", err)
	}
	if !starExists {
		t.Error("star0 does not exist but should have been created")
	}
	starExists, err = rep.Exists(testStar1.FileID, testStar1.UserID)
	if err != nil {
		t.Errorf("Failed to read whether star1 exists: %v", err)
	}
	if !starExists {
		t.Error("star1 does not exist but should have been created")
	}
}

func TestDeleteStar(t *testing.T) {
	if testStarSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testStarCleanup()
	rep := testStarSetup()
	testStarInsert(rep)

	err := rep.Delete(testStar0.FileID, testStar0.UserID)
	if err != nil {
		t.Fatalf("Failed to delete star0: %v", err)
	}

	starExists, err := rep.Exists(testStar0.FileID, testStar0.UserID)
	if err != nil {
		t.Errorf("Failed to read whether star0 exists: %v", err)
	}
	if starExists {
		t.Error("star0 does exists but should have been deleted")
	}
	starExists, err = rep.Exists(testStar1.FileID, testStar1.UserID)
	if err != nil {
		t.Errorf("Failed to read whether star1 exists: %v", err)
	}
	if !starExists {
		t.Error("star1 does not exist but should have been created and not deleted")
	}

	count, err := rep.Count()
	if err != nil {
		t.Fatalf("Failed to count stars after deletion: %v", err)
	}
	if count != 1 {
		t.Errorf("Count unequal to two after deletion: %d", count)
	}
}
