package repository

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/freecloudio/server/models"
	"github.com/freecloudio/server/utils"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	log "gopkg.in/clog.v1"
)

func init() {
	neoLabelConstraints = append(neoLabelConstraints, &neoLabelConstraint{
		label: "FSInfo",
		model: &models.FileInfo{},
	})
}

// FileInfoRepository represents a the database for storing file infos
type FileInfoRepository struct{}

// CreateFileInfoRepository creates a new FileInfoRepository IF neo4j has been inizialized
func CreateFileInfoRepository() (*FileInfoRepository, error) {
	if graphConnection == nil {
		return nil, ErrNeoNotInitialized
	}
	return &FileInfoRepository{}, nil
}

// CreateRootFolder creates the root folder for an username, does not fail if it already exists
func (rep *FileInfoRepository) CreateRootFolder(username string) (err error) {
	session, err := getGraphSession()
	if err != nil {
		return
	}
	defer session.Close()

	rootFolderInfo := &models.FileInfo{
		OwnerUsername: username,
		Name:          "/",
		LastChanged:   utils.GetTimestampNow(),
	}

	_, err = session.WriteTransaction(rep.createRootFolderTxFunc(rootFolderInfo))
	if err != nil {
		log.Error(0, "Could not create root folder for '%s': %v", username, err)
		return
	}
	return
}

func (rep *FileInfoRepository) createRootFolderTxFunc(rootFolderInfo *models.FileInfo) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		query := `
		MATCH (u:User {username: $username})
		MERGE (u)-[:HAS_ROOT_FOLDER]->(f:Folder:FSInfo {name: $fileInfo.name})
		ON CREATE SET f += $fileInfo`
		params := map[string]interface{}{
			"username": rootFolderInfo.OwnerUsername,
			"fileInfo": modelToMap(rootFolderInfo),
		}
		return tx.Run(query, params)
	}
}

// Create stores a new file info
func (rep *FileInfoRepository) Create(fileInfo *models.FileInfo) (err error) {
	session, err := getGraphSession()
	if err != nil {
		return
	}
	defer session.Close()

	_, err = session.WriteTransaction(rep.createTxFunc(fileInfo))
	if err != nil {
		log.Error(0, "Could not insert file: %v", err)
		return
	}
	return
}

func (rep *FileInfoRepository) createTxFunc(fileInfo *models.FileInfo) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		pathElements := rep.pathToElements(fileInfo.Path)
		numPathElements := len(pathElements)
		parQueryTemplate := `
			MATCH p = (u:User {username: $username})-[:HAS_ROOT_FOLDER|CONTAINS|CONTAINS_SHARED*%d]->(dir:Folder)
			WHERE [n in tail(nodes(p)) | n.name] = $pathElements
			OPTIONAL MATCH (dir)-[:CONTAINS|CONTAINS_SHARED]->(exis:FSInfo {name: $filename})
			RETURN id(dir) as parent_id, id(exis) as existing
		`
		parQuery := fmt.Sprintf(parQueryTemplate, numPathElements)
		parParams := map[string]interface{}{
			"username":     fileInfo.OwnerUsername,
			"pathElements": pathElements,
			"filename":     fileInfo.Name,
		}
		record, err := neo4j.Single(tx.Run(parQuery, parParams))
		if err != nil {
			return nil, err
		}

		if existing, ok := record.Get("existing"); !ok || existing != nil {
			return nil, errors.New("file already exists")
		}
		parentIDInt, ok := record.Get("parent_id")
		if !ok {
			return nil, errors.New("parent id not found")
		}

		insQueryTemplate := `
			MATCH (f:Folder)
			WHERE id(f) = $parentID
			CREATE (f)-[:CONTAINS]->(:%s:FSInfo $fileInfo)`
		label := "File"
		if fileInfo.IsDir {
			label = "Folder"
		}
		insQuery := fmt.Sprintf(insQueryTemplate, label)
		insParams := map[string]interface{}{
			"parentID": parentIDInt,
			"fileInfo": modelToMap(fileInfo),
		}
		return tx.Run(insQuery, insParams)
	}
}

// GetByPath returns a file info by username and path
func (rep *FileInfoRepository) GetByPath(username, path string) (fileInfo *models.FileInfo, err error) {
	session, err := getGraphSession()
	if err != nil {
		return
	}
	defer session.Close()

	fileInfoInt, err := session.ReadTransaction(rep.getByPathTxFunc(username, path))
	if err != nil {
		log.Error(0, "Could not get fileInfo for '%s' for user %s: %v", path, username, err)
		return
	}
	fileInfo = fileInfoInt.(*models.FileInfo)
	return
}

func (rep *FileInfoRepository) getByPathTxFunc(username, path string) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		pathElements := rep.pathToElements(path)
		numPathElements := len(pathElements)
		queryTemlate := `
			MATCH p = (u:User {username: $username})-[:HAS_ROOT_FOLDER|CONTAINS|CONTAINS_SHARED*%d]->(f:FSInfo)
			WHERE [n in tail(nodes(p)) | n.name] = $pathElements
			RETURN f, "Folder" IN labels(f) AS isDir
		`
		query := fmt.Sprintf(queryTemlate, numPathElements)
		params := map[string]interface{}{
			"username":     username,
			"pathElements": pathElements,
		}
		record, err := neo4j.Single(tx.Run(query, params))
		if err != nil {
			return nil, err
		}

		fileInfoInt, err := recordToModel(record, "f", &models.FileInfo{})
		if err != nil {
			return nil, err
		}
		fileInfo := fileInfoInt.(*models.FileInfo)
		isDirInt, ok := record.Get("isDir")
		if !ok {
			return nil, errors.New("isDir not part of getByPath record")
		}
		fileInfo.IsDir = isDirInt.(bool)
		fileInfo.OwnerUsername = username
		fileInfo.Path, _ = utils.SplitPath(path)

		return fileInfo, nil
	}
}

func (rep *FileInfoRepository) pathToElements(path string) []string {
	pathElements := strings.Split(path, "/")
	if len(pathElements) > 0 && pathElements[0] == "" {
		pathElements = pathElements[1:]
	}
	if length := len(pathElements); length > 0 && pathElements[length-1] == "" {
		pathElements = pathElements[:length-1]
	}
	return append([]string{"/"}, pathElements...)
}

// GetDirectoryContentByPath returns all child files of a directory
func (rep *FileInfoRepository) GetDirectoryContentByPath(username, path string) (content []*models.FileInfo, err error) {
	session, err := getGraphSession()
	if err != nil {
		return
	}
	defer session.Close()

	contentInt, err := session.ReadTransaction(rep.getDirectoryContentByPathTxFunc(username, path))
	if err != nil && IsRecordNotFoundError(err) {
		err = nil
	} else if err != nil {
		log.Error(0, "Could not get dir content for path '%s' for user '%s': %v", path, username, err)
		return
	}
	content = contentInt.([]*models.FileInfo)
	return
}

func (rep *FileInfoRepository) getDirectoryContentByPathTxFunc(username, path string) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		pathElements := rep.pathToElements(path)
		numPathElements := len(pathElements)
		queryTemplate := `
			MATCH p = (u:User {username: $username})-[:HAS_ROOT_FOLDER|CONTAINS|CONTAINS_SHARED*%d]->(dir:Folder)
			WHERE [n in tail(nodes(p)) | n.name] = $pathElements
			MATCH (dir)-[:CONTAINS|CONTAINS_SHARED]->(f:FSInfo)
			MATCH (f)<-[:CONTAINS|HAS_ROOT_FOLDER*]-(o:User)
			RETURN f, "Folder" IN labels(f) AS isDir, o.username as ownerUsername ORDER BY NOT isDir, f.name`
		query := fmt.Sprintf(queryTemplate, numPathElements)
		params := map[string]interface{}{
			"username":     username,
			"pathElements": pathElements,
		}
		res, err := tx.Run(query, params)
		if err != nil {
			return nil, err
		}

		var fileInfos []*models.FileInfo
		for res.Next() {
			fileInfoInt, err := recordToModel(res.Record(), "f", &models.FileInfo{})
			if err != nil {
				log.Error(0, "Failed to get fileInfo from record: %v", err)
				continue
			}
			fileInfo := fileInfoInt.(*models.FileInfo)
			isDirInt, ok := res.Record().Get("isDir")
			if !ok {
				log.Error(0, "isDir not part of getDirectoryContentByPath record")
				continue
			}
			fileInfo.IsDir = isDirInt.(bool)
			ownerUsernameInt, ok := res.Record().Get("ownerUsername")
			if !ok {
				log.Error(0, "ownerUsername not part of getDirectoryContentByPath record")
				continue
			}
			fileInfo.OwnerUsername = ownerUsernameInt.(string)
			fileInfo.Path = path
			fileInfos = append(fileInfos, fileInfo)
		}
		if res.Err() != nil {
			return nil, res.Err()
		}

		return fileInfos, nil
	}
}

// Delete deletes a file info by its path
func (rep *FileInfoRepository) Delete(username, path string) (err error) {
	session, err := getGraphSession()
	if err != nil {
		return
	}
	defer session.Close()

	_, err = session.WriteTransaction(rep.deleteTxFunc(username, path))
	if err != nil {
		log.Error(0, "Could not delete fileInfo: %v", err)
		return
	}
	return
}

func (rep *FileInfoRepository) deleteTxFunc(username, path string) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		pathElements := rep.pathToElements(path)
		numPathElements := len(pathElements)
		queryTemplate := `
			MATCH p = (u:User {username: $username})-[:HAS_ROOT_FOLDER|CONTAINS|CONTAINS_SHARED*%d]->(f:FSInfo)
			WHERE [n in tail(nodes(p)) | n.name] = $pathElements
			MATCH (f)-[:CONTAINS*]->(c:FSInfo)
			DETACH DELETE f, c
		`
		query := fmt.Sprintf(queryTemplate, numPathElements)
		params := map[string]interface{}{
			"username":     username,
			"pathElements": pathElements,
		}
		return tx.Run(query, params)
	}
}

// Search returns a list of file infos for a path and name search term
func (rep *FileInfoRepository) Search(username, path, term string) (results []*models.FileInfo, err error) {
	session, err := getGraphSession()
	if err != nil {
		return
	}
	defer session.Close()

	resultsInt, err := session.ReadTransaction(rep.searchTxFunc(username, path, term))
	if err != nil && IsRecordNotFoundError(err) {
		err = nil
	} else if err != nil {
		log.Error(0, "Could not get search result for term '%s' in path '%s' for user '%s': %v", term, path, username, err)
		return
	}
	results = resultsInt.([]*models.FileInfo)
	return
}

func (rep *FileInfoRepository) searchTxFunc(username, path, term string) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		pathElements := rep.pathToElements(path)
		numPathElements := len(pathElements)
		queryTemplate := `
			MATCH p = (u:User {username: $username})-[:HAS_ROOT_FOLDER|CONTAINS|CONTAINS_SHARED*%d]->(f:Folder)
			WHERE [n in tail(nodes(p)) | n.name] = $pathElements
			MATCH fp = (f)-[:CONTAINS|CONTAINS_SHARED*]->(r:FSInfo)
			WHERE toLower(r.name) CONTAINS toLower($term)
			MATCH (r)<-[:CONTAINS|HAS_ROOT_FOLDER*]-(o:User)
			RETURN r, "Folder" IN labels(r) AS isDir, o.username as ownerUsername,
				reduce(path = "", x IN nodes(fp)[1..size(nodes(fp)) - 1] | path + "/" + x.name) as relPath ORDER BY NOT isDir, r.name`
		query := fmt.Sprintf(queryTemplate, numPathElements)
		params := map[string]interface{}{
			"username":     username,
			"pathElements": pathElements,
			"term":         term,
		}
		res, err := tx.Run(query, params)
		if err != nil {
			return nil, err
		}

		var results []*models.FileInfo
		for res.Next() {
			resultInt, err := recordToModel(res.Record(), "r", &models.FileInfo{})
			if err != nil {
				log.Error(0, "Failed to get fileInfo from search record: %v", err)
				continue
			}
			result := resultInt.(*models.FileInfo)
			isDirInt, ok := res.Record().Get("isDir")
			if !ok {
				log.Error(0, "isDir not part of search record")
				continue
			}
			result.IsDir = isDirInt.(bool)
			ownerUsernameInt, ok := res.Record().Get("ownerUsername")
			if !ok {
				log.Error(0, "ownerUsername not part of search record")
				continue
			}
			result.OwnerUsername = ownerUsernameInt.(string)
			relPathInt, ok := res.Record().Get("relPath")
			if !ok {
				log.Error(0, "relPath not part of search record")
				continue
			}
			result.Path = filepath.Join(path, relPathInt.(string))
			results = append(results, result)
		}
		if res.Err() != nil {
			return nil, res.Err()
		}

		return results, nil
	}
}

// Count returns the count of file infos
func (rep *FileInfoRepository) Count() (count int64, err error) {
	session, err := getGraphSession()
	if err != nil {
		return
	}
	defer session.Close()

	countInt, err := session.ReadTransaction(rep.countTxFunc())
	if err != nil {
		log.Error(0, "Could not get count of file infos: %v", err)
		return
	}
	count = countInt.(int64)
	return
}

func (rep *FileInfoRepository) countTxFunc() neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		record, err := neo4j.Single(tx.Run("MATCH (f:FSInfo) RETURN count(*)", nil))
		if err != nil {
			return nil, err
		}

		return record.GetByIndex(0), nil
	}
}

/*
// Update updates a stored file info
func (rep *FileInfoRepository) Update(fileInfo *models.FileInfo) (err error) {
	err = sqlDatabaseConnection.Save(fileInfo).Error
	if err != nil {
		log.Error(0, "Could not update fileInfo: %v", err)
		return
	}
	return
}

// GetStarredFileInfosByUser returns all file infos a user starred
func (rep *FileInfoRepository) GetStarredFileInfosByUser(userID int64) (starredFileInfosForUser []*models.FileInfo, err error) {
	err = sqlDatabaseConnection.Raw(getStarredFilesByUserID, userID).Order(fileListOrder).Scan(&starredFileInfosForUser).Error
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

// DeleteUserFileInfos deletes all file infos for an user
func (rep *FileInfoRepository) DeleteUserFileInfos(userID int64) (err error) {
	var files []models.FileInfo
	err = sqlDatabaseConnection.Find(&files, &models.FileInfo{OwnerID: userID}).Error
	if err != nil {
		log.Error(0, "Could not get all files for %v: %v", userID, err)
		return
	}

	for _, file := range files {
		err = sqlDatabaseConnection.Delete(&file).Error
		if err != nil {
			log.Warn("Could not delete file: %v", err)
			continue
		}
	}

	return
}
*/
