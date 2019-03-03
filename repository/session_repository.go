package repository

import (
	"github.com/neo4j/neo4j-go-driver/neo4j"

	"github.com/freecloudio/server/models"
	"github.com/freecloudio/server/utils"
	log "gopkg.in/clog.v1"
)

func init() {
	neoLabelConstraints = append(neoLabelConstraints, &neoLabelConstraint{
		label: "Session",
		model: &models.Session{},
	})
}

// SessionRepository represents the database for storing sessions
type SessionRepository struct{}

// CreateSessionRepository creates a new SessionRepository IF neo4j has been initialized before
func CreateSessionRepository() (*SessionRepository, error) {
	if graphConnection == nil {
		return nil, ErrNeoNotInitialized
	}
	return &SessionRepository{}, nil
}

// Create stores a new session for an user
func (rep *SessionRepository) Create(session *models.Session, username string) (err error) {
	dbSession, err := getGraphSession()
	if err != nil {
		return
	}
	defer dbSession.Close()

	_, err = dbSession.WriteTransaction(rep.createTxFunc(session, username))
	if err != nil {
		log.Error(0, "Could not store session: %v", err)
	}
	return
}

func (rep *SessionRepository) createTxFunc(session *models.Session, username string) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		query := "MATCH (u:User {username: $username}) CREATE (u)-[:AUTHENTICATES_WITH]->(:Session $session)"
		params := map[string]interface{}{
			"session":  modelToMap(session),
			"username": username,
		}
		return tx.Run(query, params)
	}
}

// Count returns the amount of stored sessions
func (rep *SessionRepository) Count() (count int64, err error) {
	dbSession, err := getGraphSession()
	if err != nil {
		return
	}
	defer dbSession.Close()

	countInt, err := dbSession.ReadTransaction(rep.countTxFunc())
	if err != nil {
		log.Error(0, "Error counting total sessions: %v", err)
	}
	count = countInt.(int64)
	return
}

func (rep *SessionRepository) countTxFunc() neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		record, err := neo4j.Single(tx.Run("MATCH (s:Session) RETURN count(*)", nil))
		if err != nil {
			return nil, err
		}

		return record.GetByIndex(0), nil
	}
}

// Delete deletes a given session
func (rep *SessionRepository) Delete(session *models.Session) (err error) {
	dbSession, err := getGraphSession()
	if err != nil {
		return
	}
	defer dbSession.Close()

	_, err = dbSession.WriteTransaction(rep.deleteTxFunc(session.Token))
	if err != nil {
		log.Error(0, "Could not delete session: %v", err)
	}
	return
}

func (rep *SessionRepository) deleteTxFunc(sessionToken string) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		query := "MATCH (s:Session {token: $token}) DETACH DELETE s"
		params := map[string]interface{}{
			"token": sessionToken,
		}
		return tx.Run(query, params)
	}
}

// DeleteAllForUser deletes all session for one user
func (rep *SessionRepository) DeleteAllForUser(username string) (err error) {
	dbSession, err := getGraphSession()
	if err != nil {
		return
	}
	defer dbSession.Close()

	_, err = dbSession.WriteTransaction(rep.deleteAllForUserTxFunc(username))
	if err != nil {
		log.Error(0, "Could not clean all sessions for user %v: %v", username, err)
	}
	return
}

func (rep *SessionRepository) deleteAllForUserTxFunc(username string) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		query := "MATCH (s:Session)<-[:AUTHENTICATES_WITH]-(u:User {username: $username}) DETACH DELETE s"
		params := map[string]interface{}{
			"username": username,
		}
		return tx.Run(query, params)
	}
}

// DeleteExpired deletes all expired sessions
func (rep *SessionRepository) DeleteExpired() (err error) {
	log.Trace("Cleaning old sessions")
	dbSession, err := getGraphSession()
	if err != nil {
		return
	}
	defer dbSession.Close()

	currTime := utils.GetTimestampNow()
	_, err = dbSession.WriteTransaction(rep.deleteExpiredTxFunc(currTime))
	if err != nil {
		log.Error(0, "Deleting expired sessions failed: %v", err)
	}
	return
}

func (rep *SessionRepository) deleteExpiredTxFunc(currTime int64) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		query := "MATCH (s:Session) WHERE s.expires_at < $currTime DETACH DELETE s"
		params := map[string]interface{}{
			"currTime": currTime,
		}
		return tx.Run(query, params)
	}
}

// GetWithUserByToken reads and returns a session and its user by token
func (rep *SessionRepository) GetWithUserByToken(token string) (session *models.Session, user *models.User, err error) {
	dbSession, err := getGraphSession()
	if err != nil {
		return
	}
	defer dbSession.Close()

	resInt, err := dbSession.ReadTransaction(rep.getByTokenTxFunc(token))
	if err != nil {
		log.Error(0, "Could not get session by token %s: %v", token, err)
		return
	}
	res := resInt.([]interface{})
	session = res[0].(*models.Session)
	user = res[1].(*models.User)
	return
}

func (rep *SessionRepository) getByTokenTxFunc(token string) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		query := "MATCH (s:Session {token: $token})<-[:AUTHENTICATES_WITH]-(u:User) RETURN s, u"
		params := map[string]interface{}{
			"token": token,
		}
		record, err := neo4j.Single(tx.Run(query, params))
		if err != nil {
			return nil, err
		}

		session, err := recordToModel(record, "s", &models.Session{})
		if err != nil {
			return nil, err
		}
		user, err := recordToModel(record, "u", &models.User{})
		if err != nil {
			return nil, err
		}

		return []interface{}{session, user}, nil
	}
}
