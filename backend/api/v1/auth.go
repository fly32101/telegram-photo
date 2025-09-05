package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
	"github.com/telegram-photo/middleware"
	"github.com/telegram-photo/model"
)

// GitHubUser GitHub用户信息
type GitHubUser struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
}

// redirectToGitHub 重定向到GitHub授权页面
func redirectToGitHub(c *gin.Context) {
	clientID := viper.GetString("github.client_id")
	redirectURI := viper.GetString("github.redirect_uri")

	url := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=read:user",
		clientID, redirectURI,
	)

	// 如果是API请求，返回URL
	if c.GetHeader("Accept") == "application/json" {
		c.JSON(http.StatusOK, gin.H{"url": url})
		return
	}

	// 否则直接重定向
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// githubCallback GitHub授权回调
func githubCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未提供授权码"})
		return
	}

	// 获取access token
	accessToken, err := getGitHubAccessToken(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取GitHub访问令牌失败: %v", err)})
		return
	}

	// 获取用户信息
	githubUser, err := getGitHubUser(accessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取GitHub用户信息失败: %v", err)})
		return
	}

	// 创建或更新用户记录
	githubID := fmt.Sprintf("%d", githubUser.ID)
	_, err = model.CreateOrUpdateUser(githubID, githubUser.Login)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("保存用户信息失败: %v", err)})
		return
	}

	// 生成JWT令牌
	token, err := generateJWT(githubID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("生成JWT令牌失败: %v", err)})
		return
	}

	// 获取前端回调URL
	frontendCallback := viper.GetString("github.frontend_callback")
	
	// 构建重定向URL
	redirectURL := fmt.Sprintf("%s?token=%s&user_id=%s&username=%s", frontendCallback, token, githubID, githubUser.Login)
	
	// 检查Accept头，如果是JSON请求则返回JSON，否则直接重定向
	if c.GetHeader("Accept") == "application/json" {
		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user_id": githubID,
			"username": githubUser.Login,
			"redirect_url": redirectURL,
		})
	} else {
		// 直接重定向到前端回调URL
		c.Redirect(http.StatusFound, redirectURL)
	}
}

// getGitHubAccessToken 获取GitHub访问令牌
func getGitHubAccessToken(code string) (string, error) {
	clientID := viper.GetString("github.client_id")
	clientSecret := viper.GetString("github.client_secret")

	// 构建请求体
	reqBody := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          code,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// 创建请求
	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewBuffer(reqJSON))
	if err != nil {
		return "", err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 打印响应内容以便调试
	fmt.Printf("GitHub响应: %s\n", string(body))

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if token, ok := result["access_token"].(string); ok {
		return token, nil
	}

	return "", fmt.Errorf("未找到access_token: %s", string(body))
}

// getGitHubUser 获取GitHub用户信息
func getGitHubUser(accessToken string) (*GitHubUser, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// generateJWT 生成JWT令牌
func generateJWT(userID string) (string, error) {
	// 设置JWT声明
	claims := &middleware.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "telegram-photo",
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	tokenString, err := token.SignedString([]byte(viper.GetString("jwt.secret")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}