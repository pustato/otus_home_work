package cmd

import (
	"fmt"
	"io"
	"net"

	validator "github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	HTTP      HTTPConf
	GRPC      GRPCConf
	Logger    LoggerConf
	Storage   StorageConf
	Queue     QueueConf
	Scheduler SchedulerConf
}

type LoggerConf struct {
	Target   string `validate:"required"`
	Level    string `validate:"required,oneof=debug info warn error"`
	Encoding string `validate:"required,oneof=json console"`
}

type HTTPConf struct {
	Host string `validate:"required"`
	Port string `validate:"required"`
}

type GRPCConf struct {
	Host string `validate:"required"`
	Port string `validate:"required"`
}

type StorageConf struct {
	Driver     string `validate:"required,oneof=memory db"`
	DBHost     string `mapstructure:"db_host" validate:"required_if=Driver db"`
	DBPort     uint   `mapstructure:"db_port" validate:"required_if=Driver db"`
	DBUser     string `mapstructure:"db_user" validate:"required_if=Driver db"`
	DBPassword string `mapstructure:"db_password" validate:"required_if=Driver db"`
	DBName     string `mapstructure:"db_name" validate:"required_if=Driver db"`
}

type QueueConf struct {
	User     string `validate:"required"`
	Password string `validate:"required"`
	Host     string `validate:"required"`
	Port     string `validate:"required"`
	Exchange string `validate:"required"`
}

type SchedulerConf struct {
	SendNotification string `mapstructure:"send_notification" validate:"required"`
	DeleteOld        string `mapstructure:"delete_old" validate:"required"`
}

func NewConfig(r io.Reader) (*Config, error) {
	viper.SetConfigType("yml")

	setDefaults()
	bindEnv()

	if err := viper.ReadConfig(r); err != nil {
		return nil, fmt.Errorf("read in config error: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	validate := validator.New()

	if err := validate.Struct(config); err != nil {
		return nil, fmt.Errorf("config validation error: %w", err)
	}
	return &config, nil
}

func bindEnv() {
	_ = viper.BindEnv("storage.db_host", "DB_HOST")
	_ = viper.BindEnv("storage.db_port", "DB_PORT")
	_ = viper.BindEnv("storage.db_user", "DB_USER")
	_ = viper.BindEnv("storage.db_password", "DB_PASSWORD")
	_ = viper.BindEnv("storage.db_name", "DB_NAME")

	_ = viper.BindEnv("queue.user", "QUEUE_USER")
	_ = viper.BindEnv("queue.password", "QUEUE_PASSWORD")
}

func setDefaults() {
	viper.SetDefault("http.host", "0.0.0.0")
	viper.SetDefault("http.port", "8000")

	viper.SetDefault("grpc.host", "0.0.0.0")
	viper.SetDefault("grpc.port", "50051")

	viper.SetDefault("logger.target", "stderr")
	viper.SetDefault("logger.encoding", "console")

	viper.SetDefault("storage.driver", "memory")

	viper.SetDefault("queue.host", "localhost")
	viper.SetDefault("queue.port", "5672")
	viper.SetDefault("queue.exchange", "calendar")

	viper.SetDefault("scheduler.send_notification", "1m")
	viper.SetDefault("scheduler.delete_old", "0 0 */1 * *")
}

func (c *HTTPConf) Addr() string {
	return net.JoinHostPort(c.Host, c.Port)
}

func (c *GRPCConf) Addr() string {
	return net.JoinHostPort(c.Host, c.Port)
}

func (c *StorageConf) dbConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost,
		c.DBPort,
		c.DBUser,
		c.DBPassword,
		c.DBName,
	)
}

func (c *QueueConf) URI() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", c.User, c.Password, c.Host, c.Port)
}
