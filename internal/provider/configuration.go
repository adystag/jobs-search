package provider

import (
	"bytes"

	"github.com/adystag/jobs-search/internal"

	"github.com/spf13/viper"
)

type Configuration struct{}

func (Configuration) Provide(module *internal.Module) error {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()

	viper.SetDefault("DB_AUTO_MIGRATE", true)

	module.Configuration.Application.Env = viper.GetString("APP_ENV")
	module.Configuration.Application.Port = viper.GetString("APP_PORT")
	module.Configuration.Application.URL = viper.GetString("APP_URL")
	module.Configuration.Application.Secret = bytes.NewBufferString(viper.GetString("APP_SECRET")).Bytes()

	module.Configuration.JWT.LifeTime = viper.GetDuration("JWT_LIFETIME")

	module.Configuration.DB.Host = viper.GetString("DB_HOST")
	module.Configuration.DB.Port = viper.GetString("DB_PORT")
	module.Configuration.DB.User = viper.GetString("DB_USER")
	module.Configuration.DB.Password = viper.GetString("DB_PASSWORD")
	module.Configuration.DB.Name = viper.GetString("DB_NAME")
	module.Configuration.DB.AutoMigrate = viper.GetBool("DB_AUTO_MIGRATE")

	module.Configuration.DANS.BaseURL = viper.GetString("DANS_BASE_URL")

	return nil
}
