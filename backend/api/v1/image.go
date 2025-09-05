package v1

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

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

	// 上传到Telegram
	fileID, err := service.UploadImageToTelegram(file, header.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("上传图片失败: %v", err)})
		return
	}

	// 保存到数据库
	image := &model.Image{
		FileID:   fileID,
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
		"file_id":   fileID,
		"proxy_url": fmt.Sprintf("%s://%s/proxy/image/%s", c.Request.URL.Scheme, c.Request.Host, fileID),
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

	// 查询数据库
	images, total, err := model.GetImagesByUserID(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取图片列表失败: %v", err)})
		return
	}

	// 构建返回结果
	result := make([]map[string]interface{}, 0, len(images))
	for _, img := range images {
		result = append(result, map[string]interface{}{
			"id":         img.ID,
			"file_id":    img.FileID,
			"created_at": img.CreatedAt,
			"proxy_url":  fmt.Sprintf("%s/proxy/image/%s", c.Request.Host, img.FileID),
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
	// 获取用户ID
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 获取图片ID
	imageID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的图片ID"})
		return
	}

	// 查询图片是否存在且属于当前用户
	image, err := model.GetImageByID(uint(imageID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "图片不存在"})
		return
	}

	if image.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除该图片"})
		return
	}

	// 删除图片
	if err := model.DeleteImage(uint(imageID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("删除图片失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// proxyImage 代理访问图片
func proxyImage(c *gin.Context) {
	// 获取文件ID
	fileID := c.Param("file_id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未提供文件ID"})
		return
	}

	// 查询数据库验证文件ID是否存在
	_, err := model.GetImageByFileID(fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "图片不存在"})
		return
	}

	// 从Telegram获取图片
	imageURL, err := service.GetTelegramImageURL(fileID)
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
	c.Header("Content-Type", resp.Header.Get("Content-Type"))
	c.Header("Content-Length", resp.Header.Get("Content-Length"))
	c.Header("Cache-Control", "public, max-age=31536000")

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

	// 查询数据库
	images, total, err := model.GetImagesWithFilter(userID, uploadIP, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取图片列表失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"images": images,
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
