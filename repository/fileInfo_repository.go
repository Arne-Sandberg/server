package repository

import (
	"github.com/freecloudio/freecloud/models"
	log "gopkg.in/clog.v1"
)

// Add used models to enable auto migration for them
func init() {
	databaseModels = append(databaseModels, &models.FileInfo{})
}

// fileListOrder is the order in which to sort file and directory lists.
// Directories first, otherwise sorted by name.
const fileListOrder = "is_dir, name"

type FileInfoRepository struct{}

func CreateFileInfoRepository() (*FileInfoRepository, error) {
	if databaseConnection == nil {
		return nil, ErrGormNotInitialized
	}
	return &FileInfoRepository{}, nil
}

func (rep *FileInfoRepository) Create(fileInfo *models.FileInfo) (err error) {
	err = databaseConnection.Create(fileInfo).Error
	if err != nil {
		log.Error(0, "Could not insert file: %v", err)
		return
	}
	return
}

func (rep *FileInfoRepository) Delete(fileInfo *models.FileInfo) (err error) {
	err = databaseConnection.Delete(fileInfo).Error
	if err != nil {
		log.Error(0, "Could not delete fileInfo: %v", err)
		return
	}
	return
}

func (rep *FileInfoRepository) Update(fileInfo *models.FileInfo) (err error) {
	err = databaseConnection.Save(fileInfo).Error
	if err != nil {
		log.Error(0, "Could not update fileInfo: %v", err)
		return
	}
	return
}

func (rep *FileInfoRepository) GetStarredFileInfosForUser(userID int64) (starredFileInfosForUser []*models.FileInfo, err error) {
	err = databaseConnection.Where(&models.FileInfo{OwnerID: userID, Starred: true}).Order(fileListOrder).Find(&starredFileInfosForUser).Error
	if err != nil && IsRecordNotFoundError(err) {
		err = nil
		starredFileInfosForUser = make([]*models.FileInfo, 0)
	} else if err != nil {
		log.Error(0, "Could not get starred files for userID %v: %v", userID, err)
		return
	}

	return
}

func (rep *FileInfoRepository) GetSharedFileInfosForUser(userID int64) (sharedFilesForUser []*models.FileInfo, err error) {
	return
}

func (rep *FileInfoRepository) GetDirectoryContentByPath(userID int64, path, dirName string) (dirInfo *models.FileInfo, content []*models.FileInfo, err error) {
	dirInfo, err = rep.GetByPath(userID, path, dirName)
	if err != nil || !dirInfo.IsDir {
		return
	}

	content, err = rep.GetDirectoryContentByID(dirInfo.ID)
	return
}

func (rep *FileInfoRepository) GetDirectoryContentByID(directoryID int64) (content []*models.FileInfo, err error) {
	err = databaseConnection.Where(&models.FileInfo{ParentID: directoryID}).Order(fileListOrder).Find(&content).Error
	if err != nil && IsRecordNotFoundError(err) {
		err = nil
	} else if err != nil {
		log.Error(0, "Could not get dir content for dirID %v: %v", directoryID, err)
		return
	}

	return
}

func (rep *FileInfoRepository) GetByPath(userID int64, path, name string) (fileInfo *models.FileInfo, err error) {
	fileInfo = &models.FileInfo{}
	err = databaseConnection.Where(&models.FileInfo{OwnerID: userID, Path: path, Name: name}).First(fileInfo).Error
	if err != nil {
		log.Error(0, "Could not get fileInfo for %v%v for user %v: %v", path, name, userID, err)
		return
	}
	return
}

func (rep *FileInfoRepository) GetByID(fileID int64) (fileInfo *models.FileInfo, err error) {
	fileInfo = &models.FileInfo{}
	err = databaseConnection.First(fileInfo, "id = ?", fileID).Error
	if err != nil {
		log.Error(0, "Could not get fileInfo for ID %v: %v", fileID, err)
		return
	}
	return
}

func (rep *FileInfoRepository) SearchForFileInfo(userID int64, path, fileName string) (results []*models.FileInfo, err error) {
	pathSearch := path + "%"
	fileNameSearch := "%" + fileName + "%"
	err = databaseConnection.Where("owner_id = ? AND path LIKE ? AND name LIKE ?", userID, pathSearch, fileNameSearch).Order(fileListOrder).Find(&results).Error

	if err != nil && IsRecordNotFoundError(err) {
		err = nil
	} else if err != nil {
		log.Error(0, "Could not get search result for fileName %v in path %v for user %v: %v", fileName, path, userID, err)
		return
	}

	return
}

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
