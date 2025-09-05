package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Init 初始化配置
func Init() error {
	workDir, _ := os.Getwd()
	configPath := filepath.Join(workDir, "config.yaml")

	// 设置配置文件路径
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 尝试从环境变量读取配置
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		// 如果配置文件不存在，创建默认配置文件
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return createDefaultConfig(configPath)
		}
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	return nil
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig(configPath string) error {
	// 设置默认值
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("database.type", "mysql")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "3306")
	viper.SetDefault("database.user", "root")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.name", "telegram_photo")
	viper.SetDefault("telegram.bot_token", "your_telegram_bot_token")
	viper.SetDefault("telegram.chat_id", "your_telegram_chat_id")
	viper.SetDefault("github.client_id", "your_github_client_id")
	viper.SetDefault("github.client_secret", "your_github_client_secret")
	viper.SetDefault("github.redirect_uri", "http://localhost:8080/api/auth/github/callback")

	// 写入配置文件
	err := viper.WriteConfigAs(configPath)
	if err != nil {
		return fmt.Errorf("创建默认配置文件失败: %w", err)
	}

	fmt.Printf("已创建默认配置文件: %s\n", configPath)
	fmt.Println("请修改配置文件中的默认值后重新启动程序")

	return nil
}

// GetDSN 获取数据库连接字符串
func GetDSN() string {
	dbType := viper.GetString("database.type")
	host := viper.GetString("database.host")
	port := viper.GetString("database.port")
	user := viper.GetString("database.user")
	password := viper.GetString("database.password")
	dbName := viper.GetString("database.name")

	switch dbType {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			user, password, host, port, dbName)
	default:
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			user, password, host, port, dbName)
	}
}