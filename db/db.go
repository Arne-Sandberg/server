package db

import (
	"sort"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/models"
	"github.com/freecloudio/freecloud/utils"
	"github.com/pkg/errors"
	log "gopkg.in/clog.v1"
	"github.com/asdine/storm"
)

// TODO: Prevent SQL injection (docs: problem with primary key access)

type GormDB struct {
	gorm *gorm.DB
}

func NewStormDB(name string) (*GormDB, error) {
	db, err := gorm.Open("sqlite3", name)
	if err != nil {
		log.Error(0, "Could not open datbase: %v", err)
		return nil, err
	}
	log.Info("Initialized database")
	s := GormDB{gorm: db}

	db.AutoMigrate(&models.FileInfo{}, &models.User{}, &models.Session{}, &models.ShareEntry{})

	return &s, nil
}

func (db *GormDB) CleanupExpiredSessions() {
	log.Trace("Cleaning old sessions")
	var sessions []models.Session // TODO: Error handling; make nicer
	db.gorm.Find(&sessions)
	for _, sess := range sessions {
		if sess.ExpiresAt < utils.GetTimestampNow() {
			err := db.gorm.Delete(&sess)
			if err != nil {
				log.Error(0, "Deleting expired session failed: %v", err)
			}
		}
	}
}

func (db *GormDB) Close() {
	if err := db.gorm.Close(); err != nil {
		log.Fatal(0, "Error shutting down gorm: %v", err)
		return
	}

	db.gorm = nil
}

func (db *GormDB) CreateUser(user *models.User) (err error) {
	user.CreatedAt = utils.GetTimestampNow()
	user.UpdatedAt = utils.GetTimestampNow()
	err = db.gorm.Create(user).Error
	if err != nil {
		log.Error(0, "Could not create user: %v", err)
		return
	}
	return
}

func (db *GormDB) DeleteUser(userID uint32) (err error) {
	err = db.gorm.Delete(&models.User{ID: userID}).Error
	if err != nil {
		log.Error(0, "Could not delete user: %v", err)
		return
	}
	return
}

func (db *GormDB) UpdateUser(user *models.User) (err error) {
	user.UpdatedAt = utils.GetTimestampNow()
	err = db.gorm.Save(user).Error
	if err != nil {
		log.Error(0, "Could not update user: %v", err)
		return
	}
	return
}

func (db *GormDB) GetUserByID(userID uint32) (user *models.User, err error) {
	user = &models.User{}
	err = db.gorm.First(&user, userID).Error
	return
}

func (db *GormDB) GetUserByEmail(email string) (user *models.User, err error) {
	user = &models.User{}
	err = db.gorm.First(user, &models.User{Email: email}).Error
	return
}

func (db *GormDB) GetAllUsers() (users []*models.User, err error) {
	err = db.gorm.Find(&users).Error
	return
}

func (db *GormDB) GetAdminCount() (count int, err error) {
	var admins []*models.User
	err = db.gorm.Find(&admins, &models.User{IsAdmin: true}).Error
	if err != nil {
		log.Error(0, "Could not get all admins: %v", err)
		count = -1
		return
	}
	count = len(admins)
	return
}

func (db *GormDB) VerifyUserPassword(email string, plaintext string) (valid bool, err error) {
	var user models.User
	err = db.gorm.First(&user, &models.User{Email: email}).Error
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

func (db *GormDB) TotalSessionCount() (count uint32, err error) {
	err = db.gorm.Model(&models.Session{}).Count(&count).Error
	if err != nil {
		log.Error(0, "Error counting total sessions: %v", err)
		return
	}
	return
}

func (db *GormDB) StoreSession(session *models.Session) (err error) {
	err = db.gorm.Create(session).Error
	if err != nil {
		log.Error(0, "Could not store session: %v", err)
		return
	}
	return
}

func (db *GormDB) RemoveSession(session *models.Session) (err error) {
	err = db.gorm.Delete(session).Error
	if err != nil {
		log.Error(0, "Could not delete session: %v", err)
		return
	}
	return
}

func (db *GormDB) RemoveUserSessions(userID uint32) (err error) {
	var sessions []models.Session
	err = db.gorm.Find(&sessions, &models.Session{UserID: userID}).Error
	if err != nil {
		log.Error(0, "Could not get all sessions for %v: %v", userID, err)
		return
	}

	for _, session := range sessions {
		err = db.gorm.Delete(&session).Error
		if err != nil {
			log.Warn("Could not delete session: %v", err)
			return
		}
	}

	return
}

func (db *GormDB) SessionIsValid(session *models.Session) bool {
	var s models.Session
	err := db.gorm.First(&s, models.Session{Token: session.Token}).Error
	if err != nil {
		log.Info("Could not get session for verification, assuming it is invalid: %v", err)
		return false
	}

	if s.UserID != session.UserID {
		log.Warn("Session token existed, but has different UserID: %d vs %d", s.UserID, session.UserID)
		return false
	}

	now := utils.GetTimestampNow()
	log.Trace("Session expires at %v, now is %v", s.ExpiresAt, now)
	if s.ExpiresAt < now {
		log.Info("Session has expired")
		return false
	}
	return true
}

func (db *GormDB) InsertFile(fileInfo *models.FileInfo) (err error) {
	err = db.gorm.Create(fileInfo).Error
	if err != nil {
		log.Error(0, "Could not insert file: %v", err)
		return
	}
	return
}

func (db *GormDB) RemoveFile(fileInfo *models.FileInfo) (err error) {
	err = db.gorm.Delete(fileInfo).Error
	if err != nil {
		log.Error(0, "Could not delete file: %v", err)
		return
	}
	return
}

func (db *GormDB) UpdateFile(fileInfo *models.FileInfo) (err error) {
	err = db.gorm.Save(fileInfo).Error
	if err != nil {
		log.Error(0, "Could not update fileInfo: %v", err)
		return
	}
	return
}

func (db *GormDB) DeleteFile(fileInfo *models.FileInfo) (err error) {
	err = db.gorm.Delete(fileInfo).Error
	if err != nil {
		log.Error(0, "Could not delete fileInfo: %v", err)
		return
	}
	return
}

func (db *GormDB) GetStarredFilesForUser(userID uint32) (starredFilesForUser []*models.FileInfo, err error) {
	err = db.gorm.Where(&models.FileInfo{OwnerID: userID, Starred: true}).Order("isDir, name").Find(&starredFilesForUser).Error
	if err != nil && gorm.IsRecordNotFoundError(err) {
		err = nil
		starredFilesForUser = make([]*models.FileInfo, 0)
	} else if err != nil {
		log.Error(0, "Could not get starred files for userID %v: %v", userID, err)
		return
	}


	return
}

func (db *GormDB) GetSharedFilesForUser(userID uint32) (sharedFilesForUser []*models.FileInfo, err error) {
	return
}

func (db *GormDB) GetDirectoryContent(userID uint32, path, dirName string) (dirInfo *models.FileInfo, content []*models.FileInfo, err error) {
	dirInfo, err = db.GetFileInfo(userID, path, dirName)
	if err != nil || !dirInfo.IsDir {
		return
	}

	content, err = db.GetDirectoryContentWithID(dirInfo.ID)
	return
}

func (db *GormDB) GetDirectoryContentWithID(directoryID uint32) (content []*models.FileInfo, err error) {
	err = db.gorm.Where(&models.FileInfo{ParentID: directoryID}).Order("is_dir, name").Find(&content).Error
	if err != nil && gorm.IsRecordNotFoundError(err) {
		err = nil
	} else if err != nil {
		log.Error(0, "Could not get dir content for dirID %v: %v", directoryID, err)
		return
	}

	return
}

func (db *GormDB) getSortedFileInfoResultFromQuery(query storm.Query) (content []*models.FileInfo, err error) {
	err = query.OrderBy("IsDir", "Name").Find(&content)
	sort.SliceStable(content, func(i, j int) bool { return content[i].IsDir != content[j].IsDir })

	return
}

func (db *GormDB) GetFileInfo(userID uint32, path, name string) (fileInfo *models.FileInfo, err error) {
	fileInfo = &models.FileInfo{}
	err = db.gorm.Where(&models.FileInfo{OwnerID: userID, Path: path, Name: name}).First(fileInfo).Error
	if err != nil {
		log.Error(0, "Could not get fileInfo for %v%v for user %v: %v", path, name, userID, err)
		return
	}
	return
}

func (db *GormDB) GetFileInfoWithID(fileID uint32) (fileInfo *models.FileInfo, err error) {
	fileInfo = &models.FileInfo{}
	err = db.gorm.First(fileInfo, fileID).Error
	if err != nil {
		log.Error(0, "Could not get fileInfo for ID %v: %v", fileID, err)
		return
	}
	return
}

func (db *GormDB) SearchForFiles(userID uint32, path, fileName string) (results []*models.FileInfo, err error) {
	// TODO: Implement searching again
	/*pathRegex := "(?i)^" + regexp.QuoteMeta(path)
	fileNameRegex := "(?i)" + regexp.QuoteMeta(fileName)
	results, err = db.getSortedFileInfoResultFromQuery(db.gorm.Select(q.Eq("OwnerID", userID), q.Re("Path", pathRegex), q.Re("Name", fileNameRegex)))*/

	if err != nil && gorm.IsRecordNotFoundError(err) {
		err = nil
	} else if err != nil {
		log.Error(0, "Could not get search result for fileName %v in path %v for user %v: %v", fileName, path, userID, err)
		return
	}

	return
}

func (db *GormDB) DeleteUserFiles(userID uint32) (err error) {
	var files []models.FileInfo
	err = db.gorm.Find(&files, &models.FileInfo{OwnerID: userID}).Error
	if err != nil {
		log.Error(0, "Could not get all files for %v: %v", userID, err)
		return
	}

	for _, file := range files {
		err = db.gorm.Delete(&file).Error
		if err != nil {
			log.Warn("Could not delete file: %v", err)
			continue
		}
	}

	return
}

func (db *GormDB) InsertShareEntry(shareEntry *models.ShareEntry) (err error) {
	err = db.gorm.Create(shareEntry).Error
	if err != nil {
		log.Error(0, "Could not insert share entry: %v", err)
		return
	}
	return
}

func (db *GormDB) GetShareEntryByID(shareID uint32) (shareEntry *models.ShareEntry, err error) {
	shareEntry = &models.ShareEntry{}
	err = db.gorm.First(shareEntry, shareID).Error
	if err != nil {
		log.Error(0, "Could not get shareEntry for ID %v: %v", shareID, err)
		return
	}
	return
}

func (db *GormDB) GetShareEntriesForFile(fileID uint32) (shareEntries []*models.ShareEntry, err error) {
	err = db.gorm.Find(&shareEntries, &models.ShareEntry{FileID: fileID}).Error
	if err != nil {
		log.Error(0, "Could not get shareEntries for FileID %v: %v", fileID, err)
		return
	}
	return
}