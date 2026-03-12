package web

import (
	"concurrent-downloader/modules/logger"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func GetTotalSize(url string) (int64, error) {
	// 1. 记录调试信息：开始尝试获取
	logger.LogDebug("正在请求文件大小, URL: %s", url)
	n, err := GetTotalSizeWithHead(url)
	if err != nil {
		n, err = GetTotalSizeWithRange(url)
		if err != nil {
			logger.LogError("Range 请求失败: %v, URL: %s", err, url)
			return 0, err
		}
		return n, nil
	}
	return n, nil
}

func GetTotalSizeWithHead(url string) (int64, error) {
	// 1. 记录调试信息：开始尝试获取
	logger.LogDebug("正在请求文件大小, URL: %s", url)

	resp, err := http.Head(url)
	if err != nil {
		// 2. 记录错误信息：网络请求失败
		logger.LogError("HTTP请求失败: %v, URL: %s", err, url)
		return 0, err
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		logger.LogError("服务器响应异常, 状态码: %d, URL: %s", resp.StatusCode, url)
		return 0, nil // 或者根据需求返回自定义 error
	}

	// 3. 解析 Content-Length
	contentLength := resp.Header.Get("Content-Length")
	if contentLength == "" {
		logger.LogDebug("服务器未返回 Content-Length 头部, URL: %s", url)
		return 0, nil
	}

	size, err := strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		logger.LogError("解析文件大小失败: %v, Content-Length: %s", err, contentLength)
		return 0, err
	}

	// 4. 记录成功信息
	logger.LogInfo("成功获取文件大小: %d 字节, URL: %s", size, url)
	return size, nil
}

func GetTotalSizeWithRange(url string) (int64, error) {
	req, _ := http.NewRequest("GET", url, nil)

	// 关键改动：只请求第 0 字节
	req.Header.Set("Range", "bytes=0-0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// 注意：当使用 Range 时，状态码通常是 206 Partial Content
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		logger.LogError("服务器响应异常, 状态码: %d, URL: %s", resp.StatusCode, url)
		return 0, err
	}

	// 虽然只下载了 1 字节，但 Content-Length 依然能拿到（或者从 Content-Range 中解析）
	contentRange := resp.Header.Get("Content-Range")
	if contentRange == "" {
		return 0, fmt.Errorf("未找到 Content-Range 头部")
	}

	// 简单解析斜杠后面的部分
	parts := strings.Split(contentRange, "/")
	if len(parts) < 2 {
		return 0, fmt.Errorf("非法 Content-Range 格式: %s", contentRange)
	}

	size, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, err
	}

	logger.LogInfo("通过 Range 成功获取文件总大小: %d 字节", size)
	return size, nil
}
