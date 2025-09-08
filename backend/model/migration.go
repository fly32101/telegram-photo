package model

import (
	"fmt"
	"log"
)

// MigrateDatabase 执行数据库迁移
func MigrateDatabase() error {
	// 检查是否需要迁移
	var count int64
	DB.Table("information_schema.tables").Where("table_schema = DATABASE() AND table_name = 'files'").Count(&count)
	if count == 0 {
		// 文件表不存在，需要进行迁移
		return recreateTables()
	}

	return nil
}

// recreateTables 重新创建表结构
func recreateTables() error {
	log.Println("开始重新创建表结构...")

	// 检查images表是否存在
	var imagesCount int64
	DB.Table("information_schema.tables").Where("table_schema = DATABASE() AND table_name = 'images'").Count(&imagesCount)
	if imagesCount > 0 {
		// 删除images表
		log.Println("删除现有images表...")
		if err := DB.Exec("DROP TABLE IF EXISTS images;").Error; err != nil {
			return fmt.Errorf("删除images表失败: %v", err)
		}
	}

	// 创建文件表
	log.Println("创建files表...")
	if err := DB.Exec("CREATE TABLE IF NOT EXISTS `files` (" +
		"`id` bigint unsigned NOT NULL AUTO_INCREMENT," +
		"`telegram_file_id` varchar(255) NOT NULL," +
		"`md5_hash` varchar(32) NOT NULL," +
		"`created_at` datetime(3) NULL DEFAULT NULL," +
		"`updated_at` datetime(3) NULL DEFAULT NULL," +
		"`deleted_at` datetime(3) NULL DEFAULT NULL," +
		"PRIMARY KEY (`id`)," +
		"UNIQUE INDEX `idx_files_telegram_file_id` (`telegram_file_id`)," +
		"UNIQUE INDEX `idx_files_md5_hash` (`md5_hash`)," +
		"INDEX `idx_files_deleted_at` (`deleted_at`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;").Error; err != nil {
		return fmt.Errorf("创建文件表失败: %v", err)
	}

	// 创建新的images表
	log.Println("创建新的images表...")
	if err := DB.Exec("CREATE TABLE IF NOT EXISTS `images` (" +
		"`id` bigint unsigned NOT NULL AUTO_INCREMENT," +
		"`file_id` bigint unsigned NOT NULL," +
		"`user_id` varchar(255) NOT NULL," +
		"`upload_ip` varchar(255) NOT NULL," +
		"`created_at` datetime(3) NULL DEFAULT NULL," +
		"`updated_at` datetime(3) NULL DEFAULT NULL," +
		"`deleted_at` datetime(3) NULL DEFAULT NULL," +
		"PRIMARY KEY (`id`)," +
		"INDEX `idx_images_file_id` (`file_id`)," +
		"INDEX `idx_images_user_id` (`user_id`)," +
		"INDEX `idx_images_deleted_at` (`deleted_at`)," +
		"CONSTRAINT `fk_images_file` FOREIGN KEY (`file_id`) REFERENCES `files` (`id`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;").Error; err != nil {
		return fmt.Errorf("创建images表失败: %v", err)
	}

	// 创建users表
	log.Println("创建users表...")
	if err := DB.Exec("CREATE TABLE IF NOT EXISTS `users` (" +
		"`id` bigint unsigned NOT NULL AUTO_INCREMENT," +
		"`github_id` varchar(100) NOT NULL," +
		"`username` varchar(100) DEFAULT NULL," +
		"`last_login` datetime(3) DEFAULT NULL," +
		"`created_at` datetime(3) DEFAULT NULL," +
		"`updated_at` datetime(3) DEFAULT NULL," +
		"PRIMARY KEY (`id`)," +
		"UNIQUE INDEX `idx_users_github_id` (`github_id`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;").Error; err != nil {
		return fmt.Errorf("创建users表失败: %v", err)
	}

	log.Println("表结构重建完成!")
	return nil
}

// migrateToFileTable 将旧的图片数据迁移到新的文件表结构
func migrateToFileTable() error {
	log.Println("开始数据迁移: 创建文件表并迁移现有数据...")

	// 检查images表是否存在
	var imagesCount int64
	DB.Table("information_schema.tables").Where("table_schema = DATABASE() AND table_name = 'images'").Count(&imagesCount)
	if imagesCount == 0 {
		// images表不存在，直接创建新表结构
		log.Println("未检测到现有数据，创建新表结构...")
		return nil
	}

	// 检查images表结构
	var hasFileIDIndex int64
	DB.Table("information_schema.statistics").Where("table_schema = DATABASE() AND table_name = 'images' AND index_name = 'idx_images_file_id'").Count(&hasFileIDIndex)

	// 如果存在索引，先删除
	if hasFileIDIndex > 0 {
		log.Println("删除现有索引...")
		if err := DB.Exec("ALTER TABLE images DROP INDEX idx_images_file_id;").Error; err != nil {
			log.Printf("警告: 删除索引失败: %v\n", err)
		}
	}

	// 1. 创建文件表
	log.Println("创建文件表...")
	if err := DB.Exec("CREATE TABLE IF NOT EXISTS `files` (" +
		"`id` bigint unsigned NOT NULL AUTO_INCREMENT," +
		"`telegram_file_id` varchar(255) NOT NULL," +
		"`md5_hash` varchar(32) NOT NULL," +
		"`created_at` datetime(3) NULL DEFAULT NULL," +
		"`updated_at` datetime(3) NULL DEFAULT NULL," +
		"`deleted_at` datetime(3) NULL DEFAULT NULL," +
		"PRIMARY KEY (`id`)," +
		"UNIQUE INDEX `idx_files_telegram_file_id` (`telegram_file_id`)," +
		"UNIQUE INDEX `idx_files_md5_hash` (`md5_hash`)," +
		"INDEX `idx_files_deleted_at` (`deleted_at`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;").Error; err != nil {
		return fmt.Errorf("创建文件表失败: %v", err)
	}

	// 检查images表是否有md5_hash列
	var hasMD5HashColumn int64
	DB.Table("information_schema.columns").Where("table_schema = DATABASE() AND table_name = 'images' AND column_name = 'md5_hash'").Count(&hasMD5HashColumn)

	if hasMD5HashColumn > 0 {
		// 2. 从旧的图片表中提取唯一的文件信息并插入到文件表
		log.Println("迁移文件数据...")
		if err := DB.Exec("INSERT INTO files (telegram_file_id, md5_hash, created_at, updated_at) " +
			"SELECT DISTINCT file_id, md5_hash, NOW(), NOW() FROM images;").Error; err != nil {
			return fmt.Errorf("迁移文件数据失败: %v", err)
		}

		// 3. 创建临时列存储旧的file_id
		log.Println("添加临时列...")
		var hasOldFileIDColumn int64
		DB.Table("information_schema.columns").Where("table_schema = DATABASE() AND table_name = 'images' AND column_name = 'old_file_id'").Count(&hasOldFileIDColumn)
		if hasOldFileIDColumn == 0 {
			if err := DB.Exec("ALTER TABLE images ADD COLUMN old_file_id varchar(255);").Error; err != nil {
				return fmt.Errorf("添加临时列失败: %v", err)
			}
		}

		// 4. 保存旧的file_id到临时列
		log.Println("保存旧file_id...")
		if err := DB.Exec("UPDATE images SET old_file_id = file_id;").Error; err != nil {
			return fmt.Errorf("保存旧file_id失败: %v", err)
		}

		// 5. 修改images表的file_id列类型
		log.Println("修改file_id列类型...")
		if err := DB.Exec("ALTER TABLE images MODIFY COLUMN file_id bigint unsigned NOT NULL DEFAULT 1;").Error; err != nil {
			return fmt.Errorf("修改file_id列类型失败: %v", err)
		}

		// 6. 更新images表中的file_id为files表中对应的id
		log.Println("更新file_id关联...")
		if err := DB.Exec("UPDATE images i JOIN files f ON i.old_file_id = f.telegram_file_id " +
			"SET i.file_id = f.id;").Error; err != nil {
			return fmt.Errorf("更新file_id关联失败: %v", err)
		}

		// 7. 删除md5_hash列和临时列
		log.Println("删除旧列...")
		if err := DB.Exec("ALTER TABLE images DROP COLUMN md5_hash, DROP COLUMN old_file_id;").Error; err != nil {
			return fmt.Errorf("删除旧列失败: %v", err)
		}

		// 8. 添加外键约束
		log.Println("添加外键约束...")
		if err := DB.Exec("ALTER TABLE images ADD CONSTRAINT fk_images_file " +
			"FOREIGN KEY (file_id) REFERENCES files(id);").Error; err != nil {
			return fmt.Errorf("添加外键约束失败: %v", err)
		}
	} else {
		// 如果没有md5_hash列，说明表结构已经改变，但可能没有完成迁移
		log.Println("检测到表结构已部分迁移，尝试完成剩余迁移步骤...")
	}

	log.Println("数据迁移完成!")
	return nil
}