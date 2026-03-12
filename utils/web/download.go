package web

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// DownloadRange 核心底层函数：负责带重试的单块下载
// 参数 notifyProgress 是一个回调函数，每下载一点数据就会调用它，从而避开包循环引用
func DownloadRange(ctx context.Context, url string, file *os.File, start, end int64, notifyProgress func(int64)) error {
	maxRetries := 3
	var downloadedInChunk int64

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// 1. 重试退避
		if attempt > 0 {
			time.Sleep(time.Second * time.Duration(attempt))
		}

		// 2. 计算本次请求的 Range 起始点
		currentStart := start + downloadedInChunk
		if currentStart > end {
			return nil
		}

		// 3. 闭包执行单次 HTTP 请求，方便处理 defer 和错误
		err := func() error {
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				return err
			}
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", currentStart, end))

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err // 触发重试
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
				return fmt.Errorf("服务器响应异常: %d", resp.StatusCode)
			}

			// 4. 边读边写，并在写入成功后通过回调通知外部更新进度
			buf := make([]byte, 32*1024) // 32KB 缓冲区
			for {
				n, readErr := resp.Body.Read(buf)
				if n > 0 {
					// 并发安全地写入文件的绝对偏移位置
					_, writeErr := file.WriteAt(buf[:n], currentStart)
					if writeErr != nil {
						return writeErr // 磁盘错误通常不重试
					}

					// 更新局部状态
					downloadedInChunk += int64(n)
					currentStart += int64(n)

					// 5. 执行回调，把下载量传给外部的 progress 包
					if notifyProgress != nil {
						notifyProgress(int64(n))
					}
				}
				if readErr != nil {
					if readErr == io.EOF {
						return nil // 正常读取完毕
					}
					return readErr // 可能是连接断开，触发重试
				}
			}
		}()

		if err == nil {
			return nil // 下载成功，退出重试
		}
		// 如果循环继续，说明发生了 EOF 或网络错误，会进入下一次 attempt
	}
	return fmt.Errorf("分块 [%d-%d] 经过 %d 次重试后仍然失败", start, end, maxRetries)
}
