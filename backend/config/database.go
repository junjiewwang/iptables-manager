package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"iptables-management-backend/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() error {
	// 获取数据库文件路径，默认为 ./data/iptables.db
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/iptables.db"
	}

	// 确保数据库目录存在
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %v", err)
	}

	log.Printf("Attempting to connect to SQLite database: %s", dbPath)

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // SQLite可以显示更多日志
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		return fmt.Errorf("failed to connect to SQLite database: %v", err)
	}

	// 获取底层的sql.DB对象进行配置
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	// SQLite连接池配置（SQLite通常不需要太多连接）
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(1) // SQLite建议单连接
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	log.Println("SQLite database connection established successfully")

	// 启用外键约束
	if err := DB.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		log.Printf("Warning: failed to enable foreign keys: %v", err)
	}

	// 自动迁移数据库表
	err = AutoMigrate()
	if err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	// 初始化默认数据
	err = InitDefaultData()
	if err != nil {
		return fmt.Errorf("failed to initialize default data: %v", err)
	}

	log.Println("Database migrated and initialized successfully")
	return nil
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.IPTablesRule{},
		&models.OperationLog{},
	)
}

// InitDefaultData 初始化默认数据
func InitDefaultData() error {
	// 检查是否已有用户数据
	var userCount int64
	if err := DB.Model(&models.User{}).Count(&userCount).Error; err != nil {
		return fmt.Errorf("failed to count users: %v", err)
	}

	// 如果没有用户，创建默认用户
	if userCount == 0 {
		log.Println("No users found, creating default users...")

		// 创建默认管理员用户
		adminUser := models.User{
			Username: "admin",
			Password: "$2a$10$eMLPLnwGJiyqq9YHeisxneca9jbi/vc7GD9kr2Sqdrre7xmnpFsPK", // admin123
			Role:     "admin",
		}

		if err := DB.Create(&adminUser).Error; err != nil {
			return fmt.Errorf("failed to create admin user: %v", err)
		}

		// 创建默认普通用户
		normalUser := models.User{
			Username: "user1",
			Password: "$2a$10$Nu7hx3AQ2WjUQfhqQ/MAlurjoCyNAg9Ti3LWKuy8q.ILGkg.VyMZm", // user123
			Role:     "user",
		}

		if err := DB.Create(&normalUser).Error; err != nil {
			return fmt.Errorf("failed to create normal user: %v", err)
		}

		log.Println("Default users created successfully")
		log.Println("Admin user: admin/admin123")
		log.Println("Normal user: user1/user123")
	}

	return nil
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	return DB
}
