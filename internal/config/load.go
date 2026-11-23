package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

// Load reads configuration with the following precedence (lowest to highest):
// built-in defaults, config.<APP_ENV>.yaml, .env file, APP_* environment
// variables.
//
// Environment names are APP_<SECTION>_<KEY>, e.g. APP_DATABASE_DSN overrides
// database.dsn and APP_REDIS_ADDR overrides redis.addr.
func Load() (*Config, error) {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = string(ModeDev)
	}

	loadDotenv()

	v := viper.New()
	v.SetConfigName("config." + env)
	v.SetConfigType("yaml")
	v.AddConfigPath("./cmd/etc") // running from the repository root
	v.AddConfigPath(".")         // running next to the binary
	v.AddConfigPath("/etc/app")  // container-friendly absolute location

	v.SetEnvPrefix("APP")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if _, notFound := err.(viper.ConfigFileNotFoundError); !notFound {
			return nil, fmt.Errorf("config: read config.%s: %w", env, err)
		}
		// A missing file is legal: deployments may run purely on env vars.
		fmt.Printf("no config.%s.yaml found; relying on defaults and environment (env: %s)\n", env, env)
	} else {
		fmt.Printf("loaded config file: %s\n", v.ConfigFileUsed())
	}

	// viper's AutomaticEnv does not participate in Unmarshal; keys must be
	// bound explicitly. AllKeys sees every key because setDefaults seeds all
	// leaves even when the YAML file is absent.
	for _, k := range v.AllKeys() {
		v.BindEnv(k)
	}
	// Belt and braces for secrets that may be absent from both file and defaults.
	for _, k := range []string{"database.dsn", "redis.password", "redis.addr", "jwt.signing-key"} {
		v.BindEnv(k)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config: unmarshal: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	printStatus(&cfg, env)
	return &cfg, nil
}

// loadDotenv loads ./.env into the process environment without overriding
// already-set real env vars. Missing file is fine.
func loadDotenv() {
	for _, p := range []string{".env", "cmd/etc/.env"} {
		if _, err := os.Stat(p); err == nil {
			_ = gotenv.Load(p) // never overwrites already-set real env vars
			break
		}
	}
}

// printStatus is a plain stdout summary: the logger does not exist yet.
func printStatus(cfg *Config, env string) {
	fmt.Printf("app=%s version=%s env=%s mode=%s addr=%s\n",
		cfg.App.Name, cfg.App.Version, env, cfg.Server.Mode, cfg.Server.Addr)
}

// setDefaults seeds every leaf key so the AllKeys/BindEnv pass above can see
// the full schema, and so a config file only needs to carry overrides.
func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "go-web-template")
	v.SetDefault("app.version", "0.0.1")

	v.SetDefault("server.addr", ":8080")
	v.SetDefault("server.mode", "dev")
	v.SetDefault("server.read-timeout", "5s")
	v.SetDefault("server.write-timeout", "10s")
	v.SetDefault("server.shutdown-timeout", "10s")

	v.SetDefault("database.enabled", false)
	v.SetDefault("database.driver", "mysql")
	v.SetDefault("database.dsn", "")
	v.SetDefault("database.dbname", "")
	v.SetDefault("database.prometheus", true)
	v.SetDefault("database.max-open-conns", 100)
	v.SetDefault("database.max-idle-conns", 10)
	v.SetDefault("database.conn-max-lifetime", "1h")
	v.SetDefault("database.conn-max-idle-time", "30m")
	v.SetDefault("database.log-level", "warn")
	v.SetDefault("database.slow-threshold", "200ms")
	v.SetDefault("database.table-prefix", "")
	v.SetDefault("database.singular-table", true)
	v.SetDefault("database.prepare-stmt", true)
	v.SetDefault("database.skip-default-tx", false)

	v.SetDefault("redis.enabled", false)
	v.SetDefault("redis.addr", "127.0.0.1:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool-size", 100)
	v.SetDefault("redis.min-idle-conns", 10)
	v.SetDefault("redis.max-idle-conns", 100)
	v.SetDefault("redis.idle-timeout", "10m")
	v.SetDefault("redis.dial-timeout", "5s")
	v.SetDefault("redis.read-timeout", "3s")
	v.SetDefault("redis.write-timeout", "3s")
	v.SetDefault("redis.max-retries", 3)
	v.SetDefault("redis.master-name", "")
	v.SetDefault("redis.sentinel-addrs", []string{})
	v.SetDefault("redis.cluster-mode", false)

	v.SetDefault("log.level", "info")

	v.SetDefault("jwt.signing-key", "")
	v.SetDefault("jwt.issuer", "go-web-template")
	v.SetDefault("jwt.expire-duration", "24h")
	v.SetDefault("jwt.buffer-time", "30m")
	v.SetDefault("jwt.signing-method", "HS256")
	v.SetDefault("jwt.private-key-path", "")
	v.SetDefault("jwt.public-key-path", "")

	v.SetDefault("cors.enabled", true)
	v.SetDefault("cors.allow-origins", []string{"*"})
	v.SetDefault("cors.allow-methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"})
	v.SetDefault("cors.allow-headers", []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Requested-With"})
	v.SetDefault("cors.expose-headers", []string{"Content-Length", "X-Total-Count"})
	v.SetDefault("cors.allow-credentials", false)
	v.SetDefault("cors.max-age", "12h")

	v.SetDefault("prometheus.enabled", true)
	v.SetDefault("prometheus.port", 9091)
	v.SetDefault("prometheus.path", "/metrics")
	v.SetDefault("prometheus.namespace", "go-web-template")

	v.SetDefault("swagger.enabled", true)
	v.SetDefault("swagger.path", "/swagger/*any")
}
