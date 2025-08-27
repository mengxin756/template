package config

import (
    "fmt"
    "strings"

    "github.com/spf13/viper"
)

type HTTPConfig struct {
    Address string `mapstructure:"address"`
}

type LogConfig struct {
    Level       string `mapstructure:"level"`
    Encoding    string `mapstructure:"encoding"`
    Development bool   `mapstructure:"development"`
}

type Config struct {
    HTTP HTTPConfig `mapstructure:"http"`
    Log  LogConfig  `mapstructure:"log"`
    DB   DBConfig   `mapstructure:"db"`
}

type DBConfig struct {
    Driver      string `mapstructure:"driver"`
    DSN         string `mapstructure:"dsn"`
    AutoMigrate bool   `mapstructure:"auto_migrate"`
}

func Load() (*Config, error) {
    v := viper.New()
    v.SetConfigName("config")
    v.SetConfigType("yaml")
    v.AddConfigPath(".")
    v.AddConfigPath("./config")
    v.AddConfigPath("../config")
    v.AutomaticEnv()
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

    // 默认值
    v.SetDefault("http.address", ":8080")
    v.SetDefault("log.level", "info")
    v.SetDefault("log.encoding", "json")
    v.SetDefault("log.development", false)
    v.SetDefault("db.driver", "sqlite")
    v.SetDefault("db.dsn", "file:ent?mode=memory&cache=shared&_fk=1")
    v.SetDefault("db.auto_migrate", true)

    _ = v.ReadInConfig()

    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, fmt.Errorf("unmarshal config: %w", err)
    }
    return &cfg, nil
}


