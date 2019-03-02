package repository

import (
	"github.com/freecloudio/server/models"
	"github.com/freecloudio/server/utils"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	log "gopkg.in/clog.v1"
)

// UserRepository represents the database for storing users
type UserRepository struct{}

// CreateUserRepository creates a new UserRepository IF neo4j has been initialized before
func CreateUserRepository() (*UserRepository, error) {
	if graphConnection == nil {
		return nil, ErrNeoNotInitialized
	}
	return &UserRepository{}, nil
}

// Create stores a new user
func (rep *UserRepository) Create(user *models.User) (err error) {
	user.CreatedAt = utils.GetTimestampNow()
	user.UpdatedAt = utils.GetTimestampNow()

	session, err := getGraphSession()
	if err != nil {
		return
	}
	defer session.Close()

	_, err = session.WriteTransaction(rep.createTxFunc(user))
	if err != nil {
		log.Error(0, "Failed to create user: %v", err)
		return
	}
	return
}

func (rep *UserRepository) createTxFunc(user *models.User) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		return tx.Run("CREATE (:User $user)", map[string]interface{}{"user": modelToMap(user)})
	}
}

// Delete deletes a user by its username
func (rep *UserRepository) Delete(username string) (err error) {
	session, err := getGraphSession()
	if err != nil {
		return
	}
	defer session.Close()

	_, err = session.WriteTransaction(rep.deleteTxFunc(username))
	if err != nil {
		log.Error(0, "Could not delete user: %v", err)
		return
	}
	return
}

func (rep *UserRepository) deleteTxFunc(username string) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		return tx.Run("MATCH (u:User {username: $username}) DETACH DELETE u", map[string]interface{}{"username": username})
	}
}

// Update updates a user by its user.Username
func (rep *UserRepository) Update(user *models.User) (err error) {
	user.UpdatedAt = utils.GetTimestampNow()

	session, err := getGraphSession()
	if err != nil {
		return
	}
	defer session.Close()

	_, err = session.WriteTransaction(rep.updateTxFunc(user))
	if err != nil {
		log.Error(0, "Could not update user: %v", err)
		return
	}
	return
}

func (rep *UserRepository) updateTxFunc(user *models.User) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		return tx.Run("MATCH (u:User {username: $user.username}) SET u += $user", map[string]interface{}{"user": modelToMap(user)})
	}
}

// UpdateLastSession updates the 'updatedAt' and 'lastSessionAt' attributes of an user
func (rep *UserRepository) UpdateLastSession(username string) (err error) {
	session, err := getGraphSession()
	if err != nil {
		return
	}
	defer session.Close()

	currTime := utils.GetTimestampNow()
	_, err = session.WriteTransaction(rep.updateLastSessionTxFunc(username, currTime))
	if err != nil {
		log.Error(0, "Could not update users last session: %v", err)
		return
	}
	return
}

func (rep *UserRepository) updateLastSessionTxFunc(username string, currTime int64) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		return tx.Run("MATCH (u:User {username: $username}) SET u.updatedAt = $currTime, u.lastSessionAt = $currTime", map[string]interface{}{"username": username, "currTime": currTime})
	}
}

// GetByUsername reads and returns an user by username
func (rep *UserRepository) GetByUsername(username string) (user *models.User, err error) {
	session, err := getGraphSession()
	if err != nil {
		return
	}
	defer session.Close()

	userInt, err := session.ReadTransaction(rep.getByUsernameTxFunc(username))
	if err != nil {
		log.Error(0, "Failed to read user by username '%s': %v", username, err)
		return
	}
	user = userInt.(*models.User)
	return
}

func (rep *UserRepository) getByUsernameTxFunc(username string) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		record, err := neo4j.Single(tx.Run("MATCH (u:User {username: $username}) RETURN u", map[string]interface{}{"username": username}))
		if err != nil {
			return nil, err
		}

		return recordToModel(record, "u", &models.User{})
	}
}

// GetByEmail reads and return an user by email
func (rep *UserRepository) GetByEmail(email string) (user *models.User, err error) {
	session, err := getGraphSession()
	if err != nil {
		return
	}
	defer session.Close()

	userInt, err := session.ReadTransaction(rep.getByEmailTxFunc(email))
	if err != nil {
		log.Error(0, "Failed to read user by email '%s': %v", email, err)
		return
	}
	user = userInt.(*models.User)
	return
}

func (rep *UserRepository) getByEmailTxFunc(email string) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		record, err := neo4j.Single(tx.Run("MATCH (u:User {email: $email}) RETURN u", map[string]interface{}{"email": email}))
		if err != nil {
			return nil, err
		}

		return recordToModel(record, "u", &models.User{})
	}
}

// GetAll reads and returns all stored users
func (rep *UserRepository) GetAll() (users []*models.User, err error) {
	session, err := getGraphSession()
	if err != nil {
		return
	}
	defer session.Close()

	usersInt, err := session.ReadTransaction(rep.getAllTxFunc())
	if err != nil {
		log.Error(0, "Failed to read all users: %v", err)
		return
	}
	users = usersInt.([]*models.User)
	return
}

func (rep *UserRepository) getAllTxFunc() neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		res, err := tx.Run("MATCH (u:User) RETURN u", nil)
		if err != nil {
			return nil, err
		}

		var users []*models.User
		for res.Next() {
			user, err := recordToModel(res.Record(), "u", &models.User{})
			if err != nil {
				return nil, err
			}
			users = append(users, user.(*models.User))
		}
		if res.Err() != nil {
			return nil, res.Err()
		}
		return users, nil
	}
}

// AdminCount returns the amount of stored admins
func (rep *UserRepository) AdminCount() (count int64, err error) {
	session, err := getGraphSession()
	if err != nil {
		return
	}
	defer session.Close()

	countInt, err := session.ReadTransaction(rep.adminCountTxFunc())
	if err != nil {
		log.Error(0, "Could not get admin user count: %v", err)
		count = -1
		return
	}
	count = countInt.(int64)
	return
}

func (rep *UserRepository) adminCountTxFunc() neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		record, err := neo4j.Single(tx.Run("MATCH (u:User {isAdmin: true}) RETURN count(*)", nil))
		if err != nil {
			return nil, err
		}

		return record.GetByIndex(0), nil
	}
}

// TotalCount returns the amount of stored users
func (rep *UserRepository) TotalCount() (count int64, err error) {
	session, err := getGraphSession()
	if err != nil {
		return
	}
	defer session.Close()

	countInt, err := session.ReadTransaction(rep.totalCountTxFunc())
	if err != nil {
		log.Error(0, "Could not get total user count: %v", err)
		count = -1
		return
	}
	count = countInt.(int64)
	return
}

func (rep *UserRepository) totalCountTxFunc() neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		record, err := neo4j.Single(tx.Run("MATCH (u:User) RETURN count(*)", nil))
		if err != nil {
			return nil, err
		}

		return record.GetByIndex(0), nil
	}
}
