package config

// Config 定义 starter 的强类型配置入口。
type Config struct {
	App  AppConfig  `yaml:"app"`
	HTTP HTTPConfig `yaml:"http"`
	Auth AuthConfig `yaml:"auth"`
	Docs DocsConfig `yaml:"docs"`
	Log  LogConfig  `yaml:"log"`
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
