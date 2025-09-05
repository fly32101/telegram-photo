package model

import (
	"time"
)

// User 用户模型
type User struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	GitHubID  string    `gorm:"column:github_id;size:100;not null;uniqueIndex" json:"github_id"`
	Username  string    `gorm:"column:username;size:100" json:"username"`
	LastLogin time.Time `gorm:"column:last_login" json:"last_login"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// CreateOrUpdateUser 创建或更新用户
func CreateOrUpdateUser(githubID string, username string) (*User, error) {
	var user User

	// 查找用户
	result := DB.Where("github_id = ?", githubID).First(&user)

	// 如果用户不存在，创建新用户
	if result.Error != nil {
		user = User{
			GitHubID:  githubID,
			Username:  username,
			LastLogin: time.Now(),
		}
		if err := DB.Create(&user).Error; err != nil {
			return nil, err
		}
	} else {
		// 更新用户信息
		user.Username = username
		user.LastLogin = time.Now()
		if err := DB.Save(&user).Error; err != nil {
			return nil, err
		}
	}

	return &user, nil
}

// GetUserByGitHubID 根据GitHub ID获取用户
func GetUserByGitHubID(githubID string) (*User, error) {
	var user User
	if err := DB.Where("github_id = ?", githubID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
