package db

import (
	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/models"
	"github.com/freecloudio/freecloud/utils"
	"github.com/pkg/errors"
	log "gopkg.in/clog.v1"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
)

type xormDB struct {
	engine *xorm.Engine
}

var syncStructs = []interface{}{ models.FileInfo{}, models.User{}, models.Session{}, models.ShareEntry{} }

func NewXormDB(name string) (*xormDB, error) {
	engine, err := xorm.NewEngine("sqlite3", name)
	if err != nil {
		log.Error(0, "Could not open datbase: %v", err)
		return nil, err
	}
	log.Info("Initialized database")
	db := xormDB{engine: engine}

	for _, syncStruct := range syncStructs {
		err = db.engine.Sync2(syncStruct)
		log.Error(0, "Failed to sync struct '%v' in db: %v", syncStruct, err);
		return nil, err
	}

	return &db, nil
}

func (db *xormDB) CleanupExpiredSessions() {
	log.Trace("Cleaning old sessions")
	var sessions []models.Session
	err := db.engine.Find(&sessions)
	if err != nil {
		log.Error(0,"Session cleanup failed: Getting all sessions failed: %v", err)
		return
	}

	for _, sess := range sessions {
		if sess.ExpiresAt.Seconds < utils.GetTimestampNow().Seconds {
			_, err := db.engine.Delete(&sess)
			if err != nil {
				log.Error(0, "Deleting expired session failed: %v", err)
			}
		}
	}
}

func (db *xormDB) Close() {
	if err := db.engine.Close(); err != nil {
		log.Fatal(0, "Error shutting down db: %v", err)
		return
	}

	db.engine = nil
}

func (db *xormDB) CreateUser(user *models.User) (err error) {
	user.CreatedAt = utils.GetTimestampNow()
	user.UpdatedAt = utils.GetTimestampNow()
	_, err = db.engine.Insert(user)
	if err != nil {
		log.Error(0, "Could not save user: %v", err)
		return
	}
	return
}

func (db *xormDB) DeleteUser(userID uint32) (err error) {
	_, err = db.engine.Delete(&models.User{ID: userID})
	if err != nil {
		log.Error(0, "Could not delete user: %v", err)
		return
	}
	return
}

func (db *xormDB) UpdateUser(user *models.User) (err error) {
	user.UpdatedAt = utils.GetTimestampNow()
	_, err = db.engine.Update(user)
	if err != nil {
		log.Error(0, "Could not update user: %v", err)
		return
	}
	return
}

func (db *xormDB) GetUserByID(userID uint32) (user *models.User, has bool, err error) {
	user = &models.User{}
	has, err = db.engine.ID(userID).Get(user)
	return
}

func (db *xormDB) GetUserByEmail(email string) (user *models.User, has bool, err error) {
	user = &models.User{Email: email}
	has, err = db.engine.Get(user)
	return
}

func (db *xormDB) GetAllUsers() ([]*models.User, error) {
	var users []*models.User
	err := db.engine.Find(&users)
	return users, err
}

func (db *xormDB) GetAdminCount() (count int, err error) {
	var admins []*models.User
	err = db.engine.Find(&admins, &models.User{IsAdmin: true})
	if err != nil {
		log.Error(0, "Could not get all admins: %v", err)
		count = -1
		return
	}
	count = len(admins)
	return
}

func (db *xormDB) VerifyUserPassword(email string, plaintext string) (valid bool, err error) {
	user := &models.User{Email: email}
	has, err := db.engine.Get(user)
	if err != nil || !has {
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

func (db *xormDB) TotalSessionCount() uint32 {
	c, err := db.engine.Count(&models.Session{})
	if err != nil {
		log.Error(0, "Error counting total sessions: %v", err)
	}
	return uint32(c)
}

func (db *xormDB) StoreSession(session *models.Session) error {
	_, err := db.engine.Insert(session)
	return err
}

func (db *xormDB) RemoveSession(session *models.Session) error {
	_, err := db.engine.Delete(session)
	return err
}

func (db *xormDB) RemoveUserSessions(userID uint32) (err error) {
	var sessions []models.Session
	err = db.engine.Find("UserID", userID, &sessions)
	if err != nil {
		log.Error(0, "Could not get all sessions for %v: %v", userID, err)
		return
	}

	for _, session := range sessions {
		_, err = db.engine.Delete(&session)
		if err != nil {
			log.Warn("Could not delete session: %v", err)
			return
		}
	}

	return
}

func (db *xormDB) SessionIsValid(session *models.Session) bool {
	s := &models.Session{Token: session.Token}
	has, err := db.engine.Get(s)
	if err != nil || !has {
		log.Info("Could not get session for verification, assuming it is invalid: %v", err)
		return false
	}

	if s.UserID != session.UserID {
		log.Warn("Session token existed, but has different UserID: %d vs %d", s.UserID, session.UserID)
		return false
	}

	now := utils.GetTimestampNow()
	log.Trace("Session expires at %v, now is %v", s.ExpiresAt.Seconds, now.Seconds)
	if s.ExpiresAt.Seconds < now.Seconds {
		log.Info("Session has expired")
		return false
	}
	return true
}

func (db *xormDB) InsertFile(fileInfo *models.FileInfo) (err error) {
	_, err = db.engine.Insert(fileInfo)
	if err != nil {
		log.Error(0, "Could not insert file: %v", err)
		return
	}
	return
}

func (db *xormDB) RemoveFile(fileInfo *models.FileInfo) (err error) {
	_, err = db.engine.Delete(fileInfo)
	if err != nil {
		log.Error(0, "Could not delete file: %v", err)
		return
	}
	return
}

func (db *xormDB) UpdateFile(fileInfo *models.FileInfo) (err error) {
	_, err = db.engine.Insert(fileInfo)
	if err != nil {
		log.Error(0, "Could not update fileInfo: %v", err)
		return
	}
	return
}

func (db *xormDB) DeleteFile(fileInfo *models.FileInfo) (err error) {
	_, err = db.engine.Delete(fileInfo)
	if err != nil {
		log.Error(0, "Could not delete fileInfo: %v", err)
		return
	}
	return
}

func (db *xormDB) GetStarredFilesForUser(userID uint32) (starredFilesForUser []*models.FileInfo, err error) {
	err = db.engine.Asc("IsDir", "Name").Find(starredFilesForUser, &models.FileInfo{OwnerID: userID, Starred: true})
	if err == xorm.ErrNotExist { // TODO: Check error for not found from XORM
		err = nil
		starredFilesForUser = make([]*models.FileInfo, 0)
	} else if err != nil {
		log.Error(0, "Could not get starred files for userID %v: %v", userID, err)
		return
	}
	return
}

func (db *xormDB) GetSharedFilesForUser(userID uint32) (sharedFilesForUser []*models.FileInfo, err error) {
	return
}

func (db *xormDB) GetDirectoryContent(userID uint32, path, dirName string) (dirInfo *models.FileInfo, content []*models.FileInfo, err error) {
	dirInfo, err = db.GetFileInfo(userID, path, dirName)
	if err != nil || !dirInfo.IsDir {
		return
	}

	content, err = db.GetDirectoryContentWithID(dirInfo.ID)
	return
}

func (db *xormDB) GetDirectoryContentWithID(directoryID uint32) (content []*models.FileInfo, err error) {
	err = db.engine.Asc("IsDir", "Name").Find(&content, &models.FileInfo{ParentID: directoryID})

	if err == xorm.ErrNotExist { // TODO: Check error for not found from XORM
		err = nil
	} else if err != nil {
		log.Error(0, "Could not get dir content for dirID %v: %v", directoryID, err)
		return
	}

	return
}

func (db *xormDB) GetFileInfo(userID uint32, path, fileName string) (fileInfo *models.FileInfo, err error) {
	fileInfo = &models.FileInfo{}
	fileInfos := []*models.FileInfo{nil}
	err = db.engine.Find(fileInfos, &models.FileInfo{Path: path, Name: fileName, OwnerID: userID})
	if err != nil && len(fileInfos) != 1 {
		log.Error(0, "Could not get fileInfo for %v%v for user %v: %v", path, fileName, userID, err)
		return
	}

	fileInfo = fileInfos[0]
	return
}

func (db *xormDB) GetFileInfoWithID(fileID uint32) (fileInfo *models.FileInfo, err error) {
	fileInfo = &models.FileInfo{}
	_, err = db.engine.ID(fileID).Get(fileInfo)
	if err != nil {
		log.Error(0, "Could not get fileInfo for ID %v: %v", fileID, err)
		return
	}
	return
}

func (db *xormDB) SearchForFiles(userID uint32, path, fileName string) (results []*models.FileInfo, err error) {
	//res, err := db.engine.QueryInterface("select * from files where OwnerID = %d and Path like \"%s%%\" and Name like \"%%%s%%\"")
	// TODO: Parse search result into file result
	//log.Trace("Search result: %v", res)

	// TODO: Use engine.Where("a = ? AND b = ?", 1, 2).Find(&beans)
	// engine.SQL("select * from table").Find(&beans)

	if err == xorm.ErrNotExist { // TODO: Check error for not found from XORM
		err = nil
	} else if err != nil {
		log.Error(0, "Could not get search result for fileName %v in path %v for user %v: %v", fileName, path, userID, err)
		return
	}

	return
}

func (db *xormDB) DeleteUserFiles(userID uint32) (err error) {
	var files []models.FileInfo
	err = db.engine.Find("OwnerID", userID, &files)
	if err != nil {
		log.Error(0, "Could not get all files for %v: %v", userID, err)
		return
	}

	for _, file := range files {
		_, err = db.engine.Delete(&file)
		if err != nil {
			log.Warn("Could not delete file: %v", err)
			continue
		}
	}

	return
}

func (db *xormDB) InsertShareEntry(shareEntry *models.ShareEntry) (err error) {
	_, err = db.engine.Insert(shareEntry)
	if err != nil {
		log.Error(0, "Could not insert share entry: %v", err)
		return
	}
	return
}

func (db *xormDB) GetShareEntryByID(shareID uint32) (shareEntry *models.ShareEntry, err error) {
	shareEntry = &models.ShareEntry{}
	_, err = db.engine.ID(shareID).Get(shareEntry)
	if err != nil {
		log.Error(0, "Could not get shareEntry for ID %v: %v", shareID, err)
		return
	}
	return
}

func (db *xormDB) GetShareEntriesForFile(fileID uint32) (shareEntries []*models.ShareEntry, err error) {
	err = db.engine.Find(&shareEntries, &models.ShareEntry{FileID: fileID})
	if err != nil {
		log.Error(0, "Could not get shareEntries for FileID %v: %v", fileID, err)
		return
	}
	return
}