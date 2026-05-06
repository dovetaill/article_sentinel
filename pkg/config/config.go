package config

// Config 定义 starter 的强类型配置入口。
type Config struct {
	App       AppConfig       `yaml:"app"`
	HTTP      HTTPConfig      `yaml:"http"`
	Redis     RedisConfig     `yaml:"redis"`
	Database  DatabaseConfig  `yaml:"database"`
	Auth      AuthConfig      `yaml:"auth"`
	Queue     QueueConfig     `yaml:"queue"`
	Scheduler SchedulerConfig `yaml:"scheduler"`
	Docs      DocsConfig      `yaml:"docs"`
	Log       LogConfig       `yaml:"log"`
}

// AppConfig 定义应用基础监听配置。
type AppConfig struct {
	Name string `yaml:"name" env:"APP_NAME" env-required:"true"`
	Env  string `yaml:"env" env:"APP_ENV" env-default:"local"`
	Host string `yaml:"host" env:"APP_HOST" env-default:"0.0.0.0"`
	Port int    `yaml:"port" env:"APP_PORT" env-default:"8080"`
}

// HTTPConfig 定义 HTTP 服务器与应用超时配置。
type HTTPConfig struct {
	RequestTimeoutSeconds int `yaml:"request_timeout_seconds" env:"HTTP_REQUEST_TIMEOUT_SECONDS" env-default:"15"`
	ReadTimeoutSeconds    int `yaml:"read_timeout_seconds" env:"HTTP_READ_TIMEOUT_SECONDS" env-default:"15"`
	WriteTimeoutSeconds   int `yaml:"write_timeout_seconds" env:"HTTP_WRITE_TIMEOUT_SECONDS" env-default:"15"`
	IdleTimeoutSeconds    int `yaml:"idle_timeout_seconds" env:"HTTP_IDLE_TIMEOUT_SECONDS" env-default:"60"`
}

// MySQLConfig 定义主数据库为 MySQL 时的连接与连接池参数。
type MySQLConfig struct {
	Host                   string `yaml:"host" env:"DB_MYSQL_HOST"`
	Port                   int    `yaml:"port" env:"DB_MYSQL_PORT" env-default:"3306"`
	User                   string `yaml:"user" env:"DB_MYSQL_USER"`
	Password               string `yaml:"password" env:"DB_MYSQL_PASSWORD"`
	DBName                 string `yaml:"dbname" env:"DB_MYSQL_DBNAME"`
	Charset                string `yaml:"charset" env:"DB_MYSQL_CHARSET" env-default:"utf8mb4"`
	ParseTime              bool   `yaml:"parse_time" env:"DB_MYSQL_PARSE_TIME"`
	Loc                    string `yaml:"loc" env:"DB_MYSQL_LOC" env-default:"Local"`
	MaxOpenConns           int    `yaml:"max_open_conns" env:"DB_MYSQL_MAX_OPEN_CONNS" env-default:"20"`
	MaxIdleConns           int    `yaml:"max_idle_conns" env:"DB_MYSQL_MAX_IDLE_CONNS" env-default:"10"`
	ConnMaxLifetimeMinutes int    `yaml:"conn_max_lifetime_minutes" env:"DB_MYSQL_CONN_MAX_LIFETIME_MINUTES" env-default:"60"`
}

// RedisConfig 定义 Redis 连接与连接池参数。
type RedisConfig struct {
	Addr         string `yaml:"addr" env:"REDIS_ADDR" env-required:"true"`
	Password     string `yaml:"password" env:"REDIS_PASSWORD"`
	DB           int    `yaml:"db" env:"REDIS_DB" env-default:"0"`
	PoolSize     int    `yaml:"pool_size" env:"REDIS_POOL_SIZE" env-default:"10"`
	MinIdleConns int    `yaml:"min_idle_conns" env:"REDIS_MIN_IDLE_CONNS" env-default:"2"`
}

// DatabaseConfig 定义主数据库驱动与连接配置。
type DatabaseConfig struct {
	Driver   string         `yaml:"driver" env:"DB_DRIVER" env-default:"mysql"`
	MySQL    MySQLConfig    `yaml:"mysql"`
	Postgres PostgresConfig `yaml:"postgres"`
}

// PostgresConfig 定义 PostgreSQL 连接参数。
type PostgresConfig struct {
	Host                   string `yaml:"host" env:"DB_POSTGRES_HOST"`
	Port                   int    `yaml:"port" env:"DB_POSTGRES_PORT" env-default:"5432"`
	User                   string `yaml:"user" env:"DB_POSTGRES_USER"`
	Password               string `yaml:"password" env:"DB_POSTGRES_PASSWORD"`
	DBName                 string `yaml:"dbname" env:"DB_POSTGRES_DBNAME"`
	SSLMode                string `yaml:"ssl_mode" env:"DB_POSTGRES_SSL_MODE" env-default:"disable"`
	TimeZone               string `yaml:"time_zone" env:"DB_POSTGRES_TIME_ZONE" env-default:"Asia/Shanghai"`
	MaxOpenConns           int    `yaml:"max_open_conns" env:"DB_POSTGRES_MAX_OPEN_CONNS" env-default:"20"`
	MaxIdleConns           int    `yaml:"max_idle_conns" env:"DB_POSTGRES_MAX_IDLE_CONNS" env-default:"10"`
	ConnMaxLifetimeMinutes int    `yaml:"conn_max_lifetime_minutes" env:"DB_POSTGRES_CONN_MAX_LIFETIME_MINUTES" env-default:"60"`
}

// AuthConfig 定义 starter 认证配置。
type AuthConfig struct {
	Session SessionConfig `yaml:"session"`
}

// SessionConfig 定义管理台第三方跳转登录的 session 参数。
type SessionConfig struct {
	LegacySecret string `yaml:"legacy_secret" env:"AUTH_SESSION_LEGACY_SECRET"`
	Secret       string `yaml:"secret" env:"AUTH_SESSION_SECRET"`
	Issuer       string `yaml:"issuer" env:"AUTH_SESSION_ISSUER" env-default:"article-sentinel-admin"`
	TTLHours     int    `yaml:"ttl_hours" env:"AUTH_SESSION_TTL_HOURS" env-default:"24"`
	SecureCookie bool   `yaml:"secure_cookie" env:"AUTH_SESSION_SECURE_COOKIE"`
	LoginURL     string `yaml:"login_url" env:"AUTH_SESSION_LOGIN_URL" env-default:"https://appadmin.cq.qiludev.com/cq-admin/index.html"`
	RedirectURL  string `yaml:"redirect_url" env:"AUTH_SESSION_REDIRECT_URL" env-default:"/"`
}

// QueueConfig 定义后台队列运行参数。
type QueueConfig struct {
	Asynq  AsynqConfig  `yaml:"asynq"`
	Outbox OutboxConfig `yaml:"outbox"`
}

// AsynqConfig 定义 Asynq 执行器配置。
type AsynqConfig struct {
	Concurrency int    `yaml:"concurrency" env:"ASYNQ_CONCURRENCY" env-default:"10"`
	QueueName   string `yaml:"queue_name" env:"ASYNQ_QUEUE_NAME" env-default:"default"`
}

// OutboxConfig 定义持久化 outbox relay 的最小运行参数。
type OutboxConfig struct {
	Enabled                  bool   `yaml:"enabled" env:"OUTBOX_ENABLED" env-default:"true"`
	RelaySpec                string `yaml:"relay_spec" env:"OUTBOX_RELAY_SPEC" env-default:"@every 15s"`
	CleanupSpec              string `yaml:"cleanup_spec" env:"OUTBOX_CLEANUP_SPEC" env-default:"@every 1h"`
	BatchSize                int    `yaml:"batch_size" env:"OUTBOX_BATCH_SIZE" env-default:"20"`
	LeaseDurationSeconds     int    `yaml:"lease_duration_seconds" env:"OUTBOX_LEASE_DURATION_SECONDS" env-default:"30"`
	MaxAttempts              int    `yaml:"max_attempts" env:"OUTBOX_MAX_ATTEMPTS" env-default:"10"`
	DispatchedRetentionHours int    `yaml:"dispatched_retention_hours" env:"OUTBOX_DISPATCHED_RETENTION_HOURS" env-default:"168"`
	DeadLetterRetentionHours int    `yaml:"dead_letter_retention_hours" env:"OUTBOX_DEAD_LETTER_RETENTION_HOURS" env-default:"720"`
}

// SchedulerConfig 定义定时任务调度参数。
type SchedulerConfig struct {
	Enabled bool   `yaml:"enabled" env:"SCHEDULER_ENABLED" env-default:"false"`
	Spec    string `yaml:"spec" env:"SCHEDULER_SPEC" env-default:"@every 1m"`
}

// DocsConfig 定义 OpenAPI 文档开关与路径。
type DocsConfig struct {
	Enabled     bool   `yaml:"enabled" env:"DOCS_ENABLED"`
	OpenAPIPath string `yaml:"openapi_path" env:"DOCS_OPENAPI_PATH" env-default:"/openapi.json"`
	UIPath      string `yaml:"ui_path" env:"DOCS_UI_PATH" env-default:"/docs"`
}

// LogConfig 定义结构化日志输出参数。
type LogConfig struct {
	Level       string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
	Format      string `yaml:"format" env:"LOG_FORMAT" env-default:"json"`
	Output      string `yaml:"output" env:"LOG_OUTPUT" env-default:"both"`
	Dir         string `yaml:"dir" env:"LOG_DIR" env-default:"logs"`
	Filename    string `yaml:"filename" env:"LOG_FILENAME" env-default:"app.log"`
	MaxSizeMB   int    `yaml:"max_size_mb" env:"LOG_MAX_SIZE_MB" env-default:"100"`
	MaxBackups  int    `yaml:"max_backups" env:"LOG_MAX_BACKUPS" env-default:"14"`
	MaxAgeDays  int    `yaml:"max_age_days" env:"LOG_MAX_AGE_DAYS" env-default:"30"`
	Compress    bool   `yaml:"compress" env:"LOG_COMPRESS" env-default:"false"`
	RotateDaily bool   `yaml:"rotate_daily" env:"LOG_ROTATE_DAILY"`
}
