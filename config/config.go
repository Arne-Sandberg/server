package config

import "github.com/spf13/viper"

func Init() {
	viper.AutomaticEnv()
	setDefaults()
}

func setDefaults() {
	viper.SetDefault("net.host", "localhost")
	viper.SetDefault("http.port", 8080)
	viper.SetDefault("grpc.web.port", 9090)
	viper.SetDefault("grpc.nat.port", 9091)
	viper.SetDefault("grpc.nat.start", false)
	viper.SetDefault("auth.session_cookie", "fc-session")
	// Session expiry is given in hours
	viper.SetDefault("auth.session_expiry", 24)
	viper.SetDefault("http.ssl", false)
	// Upload limit is given in GB
	viper.SetDefault("http.upload_limit", 10)
	viper.SetDefault("fs.base_directory", "data")
	viper.SetDefault("fs.tmp_folder_name", ".tmp")
	viper.SetDefault("fs.tmp_data_expiry", 24)
	viper.SetDefault("db.name", "freecloud.db")
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
