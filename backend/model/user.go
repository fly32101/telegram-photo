package model

import (
	"time"
)

// User 用户模型
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	GitHubID  string    `gorm:"size:100;not null;uniqueIndex" json:"github_id"`
	Username  string    `gorm:"size:100" json:"username"`
	LastLogin time.Time `json:"last_login"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateOrUpdateUser 创建或更新用户
func CreateOrUpdateUser(githubID string, username string) (*User, error) {
	var user User

	// 查找用户
	result := DB.Where("git_hub_id = ?", githubID).First(&user)

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
