package config

import (
	"log"
	"os"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

type Config struct {
	Env        string `yaml:"env" env:"ENV" env-default:"local"`
	DSN        string `yaml:"dsn" env:"DSN"`
	GRPC       `yaml:"grpc"`
	JWT        `yaml:"jwt"`
	Mail       `yaml:"mail"`
	OTP        `yaml:"otp"`
	ResetToken `yaml:"reset_token"`
	Machinery  `yaml:"machinery"`
	Redis      `yaml:"redis"`
}

type GRPC struct {
	Port    string        `yaml:"port" env:"PORT" env-default:"9000"`
	Timeout time.Duration `yaml:"timeout" env:"TIMEOUT" env-default:"60h"`
}

type JWT struct {
	JWT_TTL int    `yaml:"ttl" env:"TTL" env-default:"60"`
	SECRET  string `yaml:"secret" env:"SECRET"`
}

type Mail struct {
	Email    string `yaml:"email" env:"MAIL_EMAIL"`
	Password string `yaml:"password" env:"MAIL_PASSWORD"`
}

type OTP struct {
	RedisName string `yaml:"redis_name" env:"OTP_REDIS_NAME"`
	OTP_TTL   int    `yaml:"ttl" env:"OTP_TTL"`
}

type ResetToken struct {
	RedisName   string `yaml:"redis_name" env:"RT_REDIS_NAME"`
	RT_TTL      int    `yaml:"ttl" env:"RT_TTL"`
	FrontendUrl string `yaml:"frontend_url" env:"RT_FRONTEND_URL"`
}

type Machinery struct {
	Broker        string `yaml:"broker" env:"MACHINERY_BROKER"`
	ResultBackend string `yaml:"result_backend" env:"MACHINERY_RESULT_BACKEND"`
}

type Redis struct {
	Address  string `yaml:"address" env:"REDIS_ADDRESS"`
	Password string `yaml:"password" env:"REDIS_PASSWORD"`
	DB       int    `yaml:"db" env:"REDIS_DB"`
}

var config *Config

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		configPath = "config/local.yml"
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file not found")
	}

	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		panic("failed to read config file: " + err.Error())
	}

	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	var cfg Config
	if err := viper.Unmarshal(&cfg, func(c *mapstructure.DecoderConfig) {
		c.TagName = "yaml"
	}); err != nil {
		panic("failed to unmarshal config: " + err.Error())
	}

	config = &cfg

	return config
}

func GetConfig() *Config {
	if config == nil {
		log.Fatal("config is not loaded")
	}
	return config
}
