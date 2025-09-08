package v1

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/telegram-photo/model"
	"github.com/telegram-photo/service"
)

// uploadImage 上传图片
func uploadImage(c *gin.Context) {
	// 获取用户ID
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 获取客户端IP
	uploadIP := c.ClientIP()

	// 获取上传的文件
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未找到上传的图片"})
		return
	}
	defer file.Close()

	// 检查文件大小
	if header.Size > 20*1024*1024 { // 20MB
		c.JSON(http.StatusBadRequest, gin.H{"error": "图片大小不能超过20MB"})
		return
	}

	// 读取文件内容用于计算MD5和上传
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("读取文件失败: %v", err)})
		return
	}

	// 计算MD5哈希
	hash := md5.Sum(fileBytes)
	md5Hash := fmt.Sprintf("%x", hash)

	// 检查是否已存在相同MD5的文件
	existingFile, err := model.GetFileByMD5Hash(md5Hash)
	var fileRecord *model.File
	var telegramFileID string
	var isExisting bool

	if err == nil && existingFile != nil {
		// 文件已存在，检查用户是否已绑定该文件
		existingImage, err := model.GetImageByFileIDAndUserID(existingFile.ID, userID)
		if err == nil && existingImage != nil {
			// 用户已绑定该文件，直接返回现有记录
			c.JSON(http.StatusOK, gin.H{
				"message":   "图片已存在",
				"file_id":   existingFile.TelegramFileID,
				"proxy_url": fmt.Sprintf("%s://%s/proxy/image/%s", c.Request.URL.Scheme, c.Request.Host, existingFile.TelegramFileID),
				"md5_hash":  md5Hash,
				"existing":  true,
			})
			return
		}

		// 用户未绑定该文件，使用现有文件记录
		fileRecord = existingFile
		telegramFileID = existingFile.TelegramFileID
		isExisting = true
	} else {
		// 文件不存在，上传到Telegram并创建新文件记录
		// 创建新的reader用于上传
		uploadReader := bytes.NewReader(fileBytes)

		// 上传到Telegram
		telegramFileID, err = service.UploadImageToTelegram(uploadReader, header.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("上传图片失败: %v", err)})
			return
		}

		// 创建文件记录
		fileRecord = &model.File{
			TelegramFileID: telegramFileID,
			MD5Hash:        md5Hash,
		}

		if err := model.CreateFile(fileRecord); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("保存文件记录失败: %v", err)})
			return
		}
		
		// 直接使用创建后的fileRecord，它应该已经有ID了
		// 如果ID为0，说明创建过程中出现了问题
		if fileRecord.ID == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建文件记录后未获取到有效ID"})
			return
		}
		
		isExisting = false
	}

	// 创建图片记录，关联用户和文件
	image := &model.Image{
		FileID:   fileRecord.ID,
		UserID:   userID,
		UploadIP: uploadIP,
	}

	if err := model.CreateImage(image); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("保存图片记录失败: %v", err)})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"message":   "上传成功",
		"file_id":   telegramFileID,
		"proxy_url": fmt.Sprintf("%s://%s/proxy/image/%s", c.Request.URL.Scheme, c.Request.Host, telegramFileID),
		"md5_hash":  md5Hash,
		"existing":  isExisting,
	})
}

// listImages 获取用户的图片列表
func listImages(c *gin.Context) {
	// 获取用户ID
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 限制页面大小
	if pageSize > 100 {
		pageSize = 100
	}

	// 查询数据库
	images, total, err := model.GetImagesByUserID(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取图片列表失败: %v", err)})
		return
	}

	// 构建返回结果
	result := make([]map[string]interface{}, 0, len(images))
	for _, img := range images {
		file, err := model.GetFileByID(img.FileID)
		if err != nil {
			continue
		}
		result = append(result, map[string]interface{}{
			"id":         img.ID,
			"file_id":    file.TelegramFileID,
			"md5_hash":   file.MD5Hash,
			"created_at": img.CreatedAt,
			"proxy_url":  fmt.Sprintf("%s://%s/proxy/image/%s", c.Request.URL.Scheme, c.Request.Host, file.TelegramFileID),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"images": result,
		"total":  total,
		"page":   page,
	})
}

// deleteImage 删除图片
func deleteImage(c *gin.Context) {
	// 获取图片ID
	imageID := c.Param("id")
	if imageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未提供图片ID"})
		return
	}

	// 转换为uint
	id, err := strconv.ParseUint(imageID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "图片ID格式错误"})
		return
	}

	// 查询数据库
	image, err := model.GetImageByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "图片不存在"})
		return
	}

	// 检查权限
	userID := c.GetString("user_id")
	isAdmin := c.GetBool("is_admin")
	if userID != image.UserID && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除该图片"})
		return
	}

	// 删除图片
	if err := model.DeleteImage(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("删除图片失败: %v", err)})
		return
	}

	// 返回成功
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// proxyImage 代理访问图片
func proxyImage(c *gin.Context) {
	// 获取文件ID
	telegramFileID := c.Param("file_id")
	if telegramFileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未提供文件ID"})
		return
	}

	// 查询数据库验证文件ID是否存在
	file, err := model.GetFileByTelegramFileID(telegramFileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "图片不存在"})
		return
	}

	// 从Telegram获取图片
	imageURL, err := service.GetTelegramImageURL(file.TelegramFileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取图片失败: %v", err)})
		return
	}

	// 获取图片内容
	resp, err := http.Get(imageURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("下载图片失败: %v", err)})
		return
	}
	defer resp.Body.Close()

	// 设置响应头
	contentType := resp.Header.Get("Content-Type")
	c.Header("Content-Type", contentType)
	c.Header("Content-Length", resp.Header.Get("Content-Length"))
	c.Header("Cache-Control", "public, max-age=31536000")

	// 根据内容类型设置不同的响应头
	if strings.HasPrefix(contentType, "image/") {
		// 如果是图片，设置为内联显示
		c.Header("Content-Disposition", "inline")
	} else {
		// 如果不是图片，设置为附件下载
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"image_%s\"", telegramFileID))
	}

	// 将图片内容写入响应
	c.Status(http.StatusOK)
	io.Copy(c.Writer, resp.Body)
}

// adminListImages 管理员获取所有图片
func adminListImages(c *gin.Context) {
	// 获取查询参数
	userID := c.Query("user_id")
	uploadIP := c.Query("upload_ip")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 限制页面大小
	if pageSize > 100 {
		pageSize = 100
	}

	// 查询数据库
	images, total, err := model.GetImagesWithFilter(userID, uploadIP, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取图片列表失败: %v", err)})
		return
	}

	// 构建响应数据
	var imageList []gin.H
	for _, img := range images {
		file, err := model.GetFileByID(img.FileID)
		if err != nil {
			continue
		}
		imageList = append(imageList, gin.H{
			"id":        img.ID,
			"file_id":   file.TelegramFileID,
			"user_id":   img.UserID,
			"upload_ip": img.UploadIP,
			"md5_hash":  file.MD5Hash,
			"created_at": img.CreatedAt,
			"proxy_url":  fmt.Sprintf("%s://%s/proxy/image/%s", c.Request.URL.Scheme, c.Request.Host, file.TelegramFileID),
		})
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"images": imageList,
		"total":  total,
		"page":   page,
	})
}

// getStats 获取统计信息
func getStats(c *gin.Context) {
	// 获取统计数据
	stats, err := model.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取统计信息失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// getImage 获取图片详情
func getImage(c *gin.Context) {
	// 获取图片ID
	imageID := c.Param("id")
	if imageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未提供图片ID"})
		return
	}

	// 转换为uint
	id, err := strconv.ParseUint(imageID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "图片ID格式错误"})
		return
	}

	// 查询数据库
	image, err := model.GetImageByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "图片不存在"})
		return
	}

	// 检查权限
	userID := c.GetString("user_id")
	isAdmin := c.GetBool("is_admin")
	if userID != image.UserID && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权访问该图片"})
		return
	}

	// 获取文件信息
	file, err := model.GetFileByID(image.FileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件信息失败"})
		return
	}

	// 返回图片信息
	c.JSON(http.StatusOK, gin.H{
		"id":        image.ID,
		"file_id":   file.TelegramFileID,
		"user_id":   image.UserID,
		"upload_ip": image.UploadIP,
		"md5_hash":  file.MD5Hash,
		"created_at": image.CreatedAt,
		"proxy_url":  fmt.Sprintf("%s://%s/proxy/image/%s", c.Request.URL.Scheme, c.Request.Host, file.TelegramFileID),
	})
}
