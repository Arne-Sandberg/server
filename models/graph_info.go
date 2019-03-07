package models

// GraphInfo holds all information about the currently connected neo4j database
type GraphInfo struct {
	Version string
	Edition string // Either 'community' or 'enterprise'
}
