package db

import (
	"regexp"
	"sort"
	"time"

	"github.com/asdine/storm/q"

	"github.com/asdine/storm"
	"github.com/asdine/storm/codec/msgpack"
	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/models"
	"github.com/pkg/errors"
	log "gopkg.in/clog.v1"
)

type StormDB struct {
	c *storm.DB
}

func NewStormDB(name string) (*StormDB, error) {
	db, err := storm.Open(name, storm.Codec(msgpack.Codec))
	if err != nil {
		log.Error(0, "Could not open datbase: %v", err)
		return nil, err
	}
	log.Info("Initialized database")
	s := StormDB{c: db}

	return &s, nil
}

func (db *StormDB) CleanupExpiredSessions() {
	log.Trace("Cleaning old sessions")
	var sessions []models.Session
	db.c.All(&sessions)
	for _, sess := range sessions {
		if time.Now().UTC().After(sess.ExpiresAt) {
			err := db.c.DeleteStruct(&sess)
			if err != nil {
				log.Error(0, "Deleting expired session failed: %v", err)
			}
		}
	}
}

func (db *StormDB) Close() {
	if err := db.c.Close(); err != nil {
		log.Fatal(0, "Error shutting down db: %v", err)
		return
	}

	db.c = nil
}

func (db *StormDB) CreateUser(user *models.User) (err error) {
	user.Created = time.Now().UTC()
	user.Updated = time.Now().UTC()
	err = db.c.Save(user)
	if err != nil {
		log.Error(0, "Could not save user: %v", err)
		return
	}
	return
}

func (db *StormDB) UpdateUser(user *models.User) (err error) {
	user.Updated = time.Now().UTC()
	err = db.c.Save(user)
	if err != nil {
		log.Error(0, "Could not update user: %v", err)
		return
	}
	return
}

func (db *StormDB) GetUserByID(uid int) (user *models.User, err error) {
	var u models.User
	err = db.c.One("ID", uid, &u)
	user = &u
	return
}

func (db *StormDB) GetUserByEmail(email string) (user *models.User, err error) {
	var u models.User
	err = db.c.One("Email", email, &u)
	user = &u
	return
}

func (db *StormDB) GetAllUsers() ([]*models.User, error) {
	var users []*models.User
	err := db.c.All(&users)
	return users, err
}

func (db *StormDB) VerifyUserPassword(email string, plaintext string) (valid bool, err error) {

	var user models.User
	err = db.c.One("Email", email, &user)
	if err != nil {
		log.Warn("Could not find user by email %s: %v", email, err)
		valid = false
		err = errors.Wrap(err, "unable to find user")
		return
	}

	// Once we got the user, verify the password
	valid, err = auth.ValidatePassword(plaintext, user.Password)
	if err != nil {
		log.Warn("Could not verify password: %v", err)
		err = errors.Wrap(err, "password verification failed")
		return
	}

	return
}

func (db *StormDB) TotalSessionCount() int {
	c, err := db.c.Count(&models.Session{})
	if err != nil {
		log.Error(0, "Error counting total sessions: %v", err)
	}
	return c
}

func (db *StormDB) StoreSession(session models.Session) error {
	return db.c.Save(&session)
}

func (db *StormDB) RemoveSession(session models.Session) error {
	return db.c.DeleteStruct(&session)
}

func (db *StormDB) SessionIsValid(session models.Session) bool {
	var s models.Session
	err := db.c.One("Token", session.Token, &s)
	if err != nil {
		log.Info("Could not get session for verification, assuming it is invalid: %v", err)
		return false
	}
	if s.UID != session.UID {
		log.Warn("Session token existed, but has different UID: %d vs %d", s.UID, session.UID)
		return false
	}
	log.Trace("Session expires at %v, now is %v", s.ExpiresAt, time.Now().UTC())
	if time.Now().UTC().After(s.ExpiresAt) {
		log.Info("Session has expired")
		return false
	}
	return true
}

func (db *StormDB) InsertFile(fileInfo *models.FileInfo) (err error) {
	err = db.c.Save(fileInfo)
	if err != nil {
		log.Error(0, "Could not insert file: %v", err)
		return
	}
	return
}

func (db *StormDB) RemoveFile(fileInfo *models.FileInfo) (err error) {
	err = db.c.DeleteStruct(fileInfo)
	if err != nil {
		log.Error(0, "Could not delete file: %v", err)
		return
	}
	return
}

func (db *StormDB) UpdateFile(fileInfo *models.FileInfo) (err error) {
	err = db.c.Save(fileInfo)
	if err != nil {
		log.Error(0, "Could not update fileInfo: %v", err)
		return
	}
	return
}

func (db *StormDB) DeleteFile(fileInfo *models.FileInfo) (err error) {
	err = db.c.DeleteStruct(fileInfo)
	if err != nil {
		log.Error(0, "Could not delete fileInfo: %v", err)
		return
	}
	return
}

func (db *StormDB) GetStarredFilesForUser(userID int) (starredFilesForUser []*models.FileInfo, err error) {
	starredFilesForUser, err = db.getSortedFileInfoResultFromQuery(db.c.Select(q.Eq("OwnerID", userID), q.Eq("Starred", true)))
	if err != nil && err.Error() == "not found" { // TODO: Is this needed? Should reference to the error directly
		err = nil
	} else if err != nil {
		log.Error(0, "Could not get starred files for userID %v: %v", userID, err)
		return
	}
	return
}

func (db *StormDB) GetDirectoryContent(userID int, path, dirName string) (dirInfo *models.FileInfo, content []*models.FileInfo, err error) {
	dirInfo, err = db.GetFileInfo(userID, path, dirName)
	if err != nil || !dirInfo.IsDir {
		return
	}

	content, err = db.GetDirectoryContentWithID(dirInfo.ID)
	return
}

func (db *StormDB) GetDirectoryContentWithID(directoryID int) (content []*models.FileInfo, err error) {
	content, err = db.getSortedFileInfoResultFromQuery(db.c.Select(q.Eq("ParentID", directoryID)))

	if err != nil && err.Error() == "not found" { // TODO: Is this needed? Should reference to the error directly
		err = nil
	} else if err != nil {
		log.Error(0, "Could not get dir content for dirID %v: %v", directoryID, err)
		return
	}

	return
}

func (db *StormDB) getSortedFileInfoResultFromQuery(query storm.Query) (content []*models.FileInfo, err error) {
	content = make([]*models.FileInfo, 0)
	err = query.OrderBy("IsDir", "Name").Find(&content)
	sort.SliceStable(content, func(i, j int) bool { return content[i].IsDir != content[j].IsDir })

	return
}

func (db *StormDB) GetFileInfo(userID int, path, fileName string) (fileInfo *models.FileInfo, err error) {
	fileInfo = &models.FileInfo{}
	err = db.c.Select(q.Eq("Path", path), q.Eq("Name", fileName), q.Eq("OwnerID", userID)).First(fileInfo)
	if err != nil {
		log.Error(0, "Could not get fileInfo for %v%v for user %v: %v", path, fileName, userID, err)
		return
	}
	return
}

func (db *StormDB) GetFileInfoWithID(fileID int) (fileInfo *models.FileInfo, err error) {
	fileInfo = &models.FileInfo{}
	err = db.c.One("ID", fileID, fileInfo)
	if err != nil {
		log.Error(0, "Could not get fileInfo for ID %v: %v", fileID, err)
		return
	}
	return
}

func (db *StormDB) SearchForFiles(userID int, path, fileName string) (results []*models.FileInfo, err error) {
	results = make([]*models.FileInfo, 0)

	pathRegex := "(?i)^" + regexp.QuoteMeta(path)
	fileNameRegex := "(?i)" + regexp.QuoteMeta(fileName)
	results, err = db.getSortedFileInfoResultFromQuery(db.c.Select(q.Eq("OwnerID", userID), q.Re("Path", pathRegex), q.Re("Name", fileNameRegex)))

	if err != nil && err.Error() == "not found" { // TODO: Is this needed? Should reference to the error directly
		err = nil
	} else if err != nil {
		log.Error(0, "Could not get search result for fileName %v in path %v for user %v: %v", fileName, path, userID, err)
		return
	}

	return
}
