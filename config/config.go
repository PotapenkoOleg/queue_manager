package config

type Config struct {
	SqlServerConfig SqlServerConfig `toml:"sql_server"`
	PostgresConfig  PostgresConfig  `toml:"postgres"`
	KafkaConfig     KafkaConfig     `toml:"kafka"`
}

type SqlServerConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Database string `toml:"database"`
	User     string `toml:"user"`
	Password string `toml:"password"`
}

type PostgresConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Database string `toml:"database"`
	User     string `toml:"user"`
	Password string `toml:"password"`
}

type KafkaConfig struct {
	Brokers    string `toml:"brokers"`
	GroupID    string `toml:"group_id"`
	ReadTopic  string `toml:"read_topic"`
	WriteTopic string `toml:"write_topic"`
}
