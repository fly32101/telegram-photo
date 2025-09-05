package model

import (
	"fmt"
	"time"

	"github.com/telegram-photo/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Init 初始化数据库连接
func Init() error {
	dsn := config.GetDSN()
	var err error

	// 连接数据库
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 自动迁移表结构
	err = DB.AutoMigrate(&Image{}, &User{})
	if err != nil {
		return fmt.Errorf("迁移数据表失败: %w", err)
	}

	return nil
}

// Image 图片模型
type Image struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FileID    string    `gorm:"size:255;not null;uniqueIndex" json:"file_id"`
	UserID    string    `gorm:"size:100;not null;index" json:"user_id"`
	UploadIP  string    `gorm:"size:50" json:"upload_ip"`
	MD5Hash   string    `gorm:"size:32;uniqueIndex" json:"md5_hash"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateImage 创建图片记录
func CreateImage(image *Image) error {
	return DB.Create(image).Error
}

// GetImageByFileID 根据FileID获取图片
func GetImageByFileID(fileID string) (*Image, error) {
	var image Image
	err := DB.Where("file_id = ?", fileID).First(&image).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

// GetImageByMD5Hash 根据MD5哈希获取图片
func GetImageByMD5Hash(md5Hash string) (*Image, error) {
	var image Image
	err := DB.Where("md5_hash = ?", md5Hash).First(&image).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

// GetImageByID 根据ID获取图片
func GetImageByID(id uint) (*Image, error) {
	var image Image
	err := DB.First(&image, id).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

// GetImagesByUserID 获取用户的所有图片
func GetImagesByUserID(userID string, page, pageSize int) ([]Image, int64, error) {
	var images []Image
	var total int64

	// 获取总数
	DB.Model(&Image{}).Where("user_id = ?", userID).Count(&total)

	// 分页查询
	err := DB.Where("user_id = ?", userID).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&images).Error

	if err != nil {
		return nil, 0, err
	}

	return images, total, nil
}

// DeleteImage 删除图片
func DeleteImage(id uint) error {
	return DB.Delete(&Image{}, id).Error
}

// GetImagesWithFilter 根据条件筛选图片
func GetImagesWithFilter(userID, uploadIP string, page, pageSize int) ([]Image, int64, error) {
	var images []Image
	var total int64
	query := DB.Model(&Image{})

	// 应用筛选条件
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if uploadIP != "" {
		query = query.Where("upload_ip = ?", uploadIP)
	}

	// 获取总数
	query.Count(&total)

	// 分页查询
	err := query.Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&images).Error

	if err != nil {
		return nil, 0, err
	}

	return images, total, nil
}

// GetStats 获取统计信息
func GetStats() (map[string]interface{}, error) {
	var totalImages int64
	var todayImages int64
	var userCount int64

	// 获取总图片数
	DB.Model(&Image{}).Count(&totalImages)

	// 获取今日上传数
	today := time.Now().Format("2006-01-02")
	DB.Model(&Image{}).
		Where("DATE(created_at) = ?", today).
		Count(&todayImages)

	// 获取用户数
	DB.Model(&Image{}).
		Distinct("user_id").
		Count(&userCount)

	// 获取用户上传排行
	type UserStat struct {
		UserID string
		Count  int
	}

	var userStats []UserStat
	DB.Model(&Image{}).
		Select("user_id, COUNT(*) as count").
		Group("user_id").
		Order("count DESC").
		Limit(10).
		Scan(&userStats)

	return map[string]interface{}{
		"total_images":  totalImages,
		"today_images":  todayImages,
		"user_count":    userCount,
		"user_rankings": userStats,
	}, nil
}