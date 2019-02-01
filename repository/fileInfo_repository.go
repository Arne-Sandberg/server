package repository

import (
	"github.com/freecloudio/server/models"
	log "gopkg.in/clog.v1"
)

// Add used models to enable auto migration for them
func init() {
	databaseModels = append(databaseModels, &models.FileInfo{})
}

// fileListOrder is the order in which to sort file and directory lists.
// Directories first, otherwise sorted by name.
const fileListOrder = "is_dir, name"

// FileInfoRepository represents a the database for storing file infos
type FileInfoRepository struct{}

// CreateFileInfoRepository creates a new FileInfoRepository IF gorm has been inizialized
func CreateFileInfoRepository() (*FileInfoRepository, error) {
	if databaseConnection == nil {
		return nil, ErrGormNotInitialized
	}
	return &FileInfoRepository{}, nil
}

// Create stores a new file info
func (rep *FileInfoRepository) Create(fileInfo *models.FileInfo) (err error) {
	err = databaseConnection.Create(fileInfo).Error
	if err != nil {
		log.Error(0, "Could not insert file: %v", err)
		return
	}
	return
}

// Delete deletes a file info by its fileInfoID
func (rep *FileInfoRepository) Delete(fileInfoID int64) (err error) {
	err = databaseConnection.Delete(&models.FileInfo{ID: fileInfoID}).Error
	if err != nil {
		log.Error(0, "Could not delete fileInfo: %v", err)
		return
	}
	return
}

// Update updates a stored file info
func (rep *FileInfoRepository) Update(fileInfo *models.FileInfo) (err error) {
	err = databaseConnection.Save(fileInfo).Error
	if err != nil {
		log.Error(0, "Could not update fileInfo: %v", err)
		return
	}
	return
}

// GetStarredFileInfosByUser returns all file infos a user starred
func (rep *FileInfoRepository) GetStarredFileInfosByUser(userID int64) (starredFileInfosForUser []*models.FileInfo, err error) {
	err = databaseConnection.Raw(getStarredFilesByUserID, userID).Order(fileListOrder).Scan(&starredFileInfosForUser).Error
	if err != nil && IsRecordNotFoundError(err) {
		err = nil
		starredFileInfosForUser = make([]*models.FileInfo, 0)
	} else if err != nil {
		log.Error(0, "Could not get starred files for userID %v: %v", userID, err)
		return
	}

	return
}

// GetSharedWithFileInfosByUser returns all file infos shared with the user
func (rep *FileInfoRepository) GetSharedWithFileInfosByUser(userID int64) (sharedFilesForUser []*models.FileInfo, err error) {
	return
}

// GetSharedFileInfosByUser returns all file infos a user shared with someone else
func (rep *FileInfoRepository) GetSharedFileInfosByUser(userID int64) (sharedFilesForUser []*models.FileInfo, err error) {
	return
}

// GetDirectoryContentByID returns all direct child files of a directory with stars for an user; for no stars use userID '0'
func (rep *FileInfoRepository) GetDirectoryContentByID(directoryID, userID int64) (content []*models.FileInfo, err error) {
	if userID > 0 {
		err = databaseConnection.Raw(getDirectoryContent, directoryID, userID).Order(fileListOrder).Scan(&content).Error
	} else {
		err = databaseConnection.Where(&models.FileInfo{ParentID: directoryID}).Order(fileListOrder).Order(fileListOrder).Find(&content).Error
	}
	if err != nil && IsRecordNotFoundError(err) {
		err = nil
	} else if err != nil {
		log.Error(0, "Could not get dir content for dirID %v for user %v: %v", directoryID, userID, err)
		return
	}

	return
}

// GetByPath returns a file info by userID, path and name AND the owner is the user
func (rep *FileInfoRepository) GetByPath(userID int64, path, name string) (fileInfo *models.FileInfo, err error) {
	fileInfo = &models.FileInfo{}
	err = databaseConnection.Raw(getByPathForOwner, path, name, userID, userID).Scan(fileInfo).Error
	if err != nil {
		log.Error(0, "Could not get fileInfo for %v%v for user %v: %v", path, name, userID, err)
		return
	}
	return
}

// GetByID returns a file by its fileID AND the owner is the user
func (rep *FileInfoRepository) GetByID(fileID int64) (fileInfo *models.FileInfo, err error) {
	fileInfo = &models.FileInfo{}
	err = databaseConnection.First(fileInfo, "id = ?", fileID).Error
	if err != nil {
		log.Error(0, "Could not get fileInfo for ID %v: %v", fileID, err)
		return
	}
	return
}

// SearchForFileInfo returns a list of file infos for a path and name search term
// TODO: With starred
func (rep *FileInfoRepository) SearchForFileInfo(userID int64, path, name string) (results []*models.FileInfo, err error) {
	pathSearch := path + "%"
	fileNameSearch := "%" + name + "%"
	err = databaseConnection.Where("owner_id = ? AND path LIKE ? AND name LIKE ?", userID, pathSearch, fileNameSearch).Order(fileListOrder).Find(&results).Error

	if err != nil && IsRecordNotFoundError(err) {
		err = nil
	} else if err != nil {
		log.Error(0, "Could not get search result for fileName %v in path %v for user %v: %v", name, path, userID, err)
		return
	}

	return
}

// DeleteUserFileInfos deletes all file infos for an user
func (rep *FileInfoRepository) DeleteUserFileInfos(userID int64) (err error) {
	var files []models.FileInfo
	err = databaseConnection.Find(&files, &models.FileInfo{OwnerID: userID}).Error
	if err != nil {
		log.Error(0, "Could not get all files for %v: %v", userID, err)
		return
	}

	for _, file := range files {
		err = databaseConnection.Delete(&file).Error
		if err != nil {
			log.Warn("Could not delete file: %v", err)
			continue
		}
	}

	return
}

// Count returns the count of file infos
func (rep *FileInfoRepository) Count() (count int64, err error) {
	err = databaseConnection.Model(&models.FileInfo{}).Count(&count).Error
	if err != nil {
		log.Error(0, "Could not get count of file infos: %v", err)
		return
	}
	return
}

var (
	selectPart    = "select file.id, file.is_dir, file.last_changed, file.mime_type, file.name, file.owner_id, file.parent_id, file.path, file.share_id, file.size, (stars.file_id is not null) as starred"
	joinStarsPart = " left outer join stars on stars.file_id = file.id and stars.user_id = ?"

	getStarredFilesByUserID = selectPart + " from file_infos as file join stars on stars.file_id = file.id where stars.user_id = ?"                                       // Only userID
	getDirectoryContent     = selectPart + " from (select * from file_infos as file where file.parent_id = ?) as file" + joinStarsPart                                    // ParentID and userID
	getByPathForOwner       = selectPart + " from (select * from file_infos as file where file.path = ? and file.name = ? and file.owner_id = ?) as file" + joinStarsPart // Path, name and two times userID
)
