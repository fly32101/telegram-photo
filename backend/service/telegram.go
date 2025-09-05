package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	telegramAPIBaseURL = "https://api.telegram.org/bot%s"
	telegramFileBaseURL = "https://api.telegram.org/file/bot%s/%s"
)

// TelegramResponse Telegram API响应
type TelegramResponse struct {
	Ok          bool            `json:"ok"`
	Result      json.RawMessage `json:"result"`
	ErrorCode   int             `json:"error_code,omitempty"`
	Description string          `json:"description,omitempty"`
}

// PhotoSize Telegram照片尺寸
type PhotoSize struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	FileSize     int    `json:"file_size,omitempty"`
}

// Message Telegram消息
type Message struct {
	MessageID int         `json:"message_id"`
	Photo     []PhotoSize `json:"photo,omitempty"`
	Document  *Document   `json:"document,omitempty"`
}

// Document Telegram文档
type Document struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	FileName     string `json:"file_name,omitempty"`
	MimeType     string `json:"mime_type,omitempty"`
	FileSize     int    `json:"file_size,omitempty"`
}

// File Telegram文件
type File struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	FileSize     int    `json:"file_size,omitempty"`
	FilePath     string `json:"file_path,omitempty"`
}

// UploadImageToTelegram 上传图片到Telegram
func UploadImageToTelegram(file io.Reader, filename string) (string, error) {
	// 获取配置
	botToken := viper.GetString("telegram.bot_token")
	chatID := viper.GetString("telegram.chat_id")

	if botToken == "" || chatID == "" {
		return "", fmt.Errorf("Telegram配置不完整")
	}

	// 准备请求URL
	apiURL := fmt.Sprintf(telegramAPIBaseURL+"/sendPhoto", botToken)

	// 创建multipart请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加chat_id字段
	if err := writer.WriteField("chat_id", chatID); err != nil {
		return "", err
	}

	// 添加图片文件
	part, err := writer.CreateFormFile("photo", filepath.Base(filename))
	if err != nil {
		return "", err
	}

	// 复制文件内容
	if _, err := io.Copy(part, file); err != nil {
		return "", err
	}

	// 完成multipart写入
	if err := writer.Close(); err != nil {
		return "", err
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", apiURL, body)
	if err != nil {
		return "", err
	}

	// 设置Content-Type
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 解析响应
	var telegramResp TelegramResponse
	if err := json.NewDecoder(resp.Body).Decode(&telegramResp); err != nil {
		return "", err
	}

	// 检查响应状态
	if !telegramResp.Ok {
		return "", fmt.Errorf("Telegram API错误: %s", telegramResp.Description)
	}

	// 解析消息
	var message Message
	if err := json.Unmarshal(telegramResp.Result, &message); err != nil {
		return "", err
	}

	// 获取文件ID
	var fileID string
	if len(message.Photo) > 0 {
		// 使用最大尺寸的图片
		fileID = message.Photo[len(message.Photo)-1].FileID
	} else if message.Document != nil {
		fileID = message.Document.FileID
	} else {
		return "", fmt.Errorf("未找到上传的文件ID")
	}

	return fileID, nil
}

// GetTelegramImageURL 获取Telegram图片URL
func GetTelegramImageURL(fileID string) (string, error) {
	// 获取配置
	botToken := viper.GetString("telegram.bot_token")

	if botToken == "" {
		return "", fmt.Errorf("Telegram配置不完整")
	}

	// 准备请求URL
	apiURL := fmt.Sprintf(telegramAPIBaseURL+"/getFile", botToken)
	apiURL = fmt.Sprintf("%s?file_id=%s", apiURL, fileID)

	// 发送请求
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 解析响应
	var telegramResp TelegramResponse
	if err := json.NewDecoder(resp.Body).Decode(&telegramResp); err != nil {
		return "", err
	}

	// 检查响应状态
	if !telegramResp.Ok {
		return "", fmt.Errorf("Telegram API错误: %s", telegramResp.Description)
	}

	// 解析文件信息
	var file File
	if err := json.Unmarshal(telegramResp.Result, &file); err != nil {
		return "", err
	}

	// 检查文件路径
	if file.FilePath == "" {
		return "", fmt.Errorf("未找到文件路径")
	}

	// 构建文件URL
	fileURL := fmt.Sprintf(telegramFileBaseURL, botToken, file.FilePath)

	return fileURL, nil
}