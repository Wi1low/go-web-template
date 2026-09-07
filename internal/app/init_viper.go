package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/Wi1low/go-web-template/internal/config"
	"github.com/spf13/viper"
)

var AppConfig *config.Config

func LoadConfig() error {
	viper.Reset()

	// 1. 获取系统环境变量 APP_ENV，如果没设置，默认当做 dev 环境
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	configName := fmt.Sprintf("config.%s", env)
	if env == "release" {
		configName = "config.prod"
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")

	// 指定配置文件可能存放的目录
	viper.AddConfigPath("./cmd/etc") // 开发时通常在项目根目录运行
	viper.AddConfigPath(".")         // 部署时通常和二进制文件在同一级

	// 3. 环境变量覆盖机制 (非常重要，用于读取 Docker 和 GitHub Secrets)
	viper.SetEnvPrefix("APP")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// 4. 设置默认值
	setDefaults()

	// 5. 执行读取
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("未找到配置文件，将完全依赖系统环境变量和默认值 (当前环境: %s)\n", env)
		} else {
			return fmt.Errorf("读取配置文件发生致命错误: %w", err)
		}
	}

	// 6. 绑定到结构体
	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("配置解析到结构体失败: %w", err)
	}
	AppConfig = &cfg

	// 打印状态，方便在控制台确认加载了哪个文件
	printConfigStatus(cfg, env)

	return nil
}

func printConfigStatus(cfg config.Config, env string) {
	fmt.Println("\n=== 配置加载结果 ===")
	fmt.Printf("运行环境 (APP_ENV): %s\n", env)
	fmt.Printf("实际使用的配置文件: %s\n", viper.ConfigFileUsed())
	fmt.Printf("应用名称: %s\n", cfg.App.Name)
	fmt.Printf("服务器模式: %s\n", cfg.Server.Mode)
	fmt.Println("====================")
}

func setDefaults() {
	viper.SetDefault("server.port", ":8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("database.max-open-conns", 100)
	viper.SetDefault("log.level", "info")
}
