package config

import "github.com/spf13/viper"

func init() {
	viper.SetEnvPrefix("FC")
	viper.AutomaticEnv()
	setDefaults()
}

func setDefaults() {
	viper.SetDefault("net.host", "localhost")

	viper.SetDefault("http.port", 8080)
	viper.SetDefault("http.ssl", false)
	// Upload limit is given in GB
	viper.SetDefault("http.upload_limit", 10)

	viper.SetDefault("auth.session_cookie", "fc-session")
	// Session expiry and cleanup interval are given in hours
	viper.SetDefault("auth.session_expiry", 24)
	viper.SetDefault("auth.session_cleanup_interval", 1)

	viper.SetDefault("fs.base_directory", "data")
	viper.SetDefault("fs.avatar_directory", "avatars")
	viper.SetDefault("fs.tmp_clear_interval", 6)
	viper.SetDefault("fs.tmp_data_expiry", 24)

	viper.SetDefault("db.type", "sqlite3")
	viper.SetDefault("db.host", "")
	viper.SetDefault("db.port", 0)
	viper.SetDefault("db.user", "")
	viper.SetDefault("db.password", "")
	viper.SetDefault("db.name", "freecloud.db")

	viper.SetDefault("graph_url", "bolt://localhost:7687")
	viper.SetDefault("graph_user", "neo4j")
	viper.SetDefault("graph_password", "freecloud")
	viper.SetDefault("graph_enterprise", false)
}

func GetString(key string) string {
	return viper.GetString(key)
}

func GetInt(key string) int {
	return viper.GetInt(key)
}

func GetInt64(key string) int64 {
	return viper.GetInt64(key)
}

func GetBool(key string) bool {
	return viper.GetBool(key)
}
