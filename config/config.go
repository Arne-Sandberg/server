package config

import "github.com/spf13/viper"

func Init() {
	viper.AutomaticEnv()
	setDefaults()
}

func setDefaults() {
	viper.SetDefault("http.host", "localhost")
	viper.SetDefault("http.port", 8080)
	viper.SetDefault("auth.session_cookie", "fc-session")
	viper.SetDefault("auth.user_cookie", "fc-user")
	viper.SetDefault("http.ssl", false)
	viper.SetDefault("http.upload_limit", 10)
	viper.SetDefault("fs.base_directory", "data")
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
