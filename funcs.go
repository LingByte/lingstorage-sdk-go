package lingstorage

import (
	"fmt"
	"time"
)

type ProgressMonitor struct {
	startTime    time.Time
	lastTime     time.Time
	lastUploaded int64
	callback     func(uploaded, total int64, percentage float64, speed float64, eta string)
}

func NewProgressMonitor() *ProgressMonitor {
	return &ProgressMonitor{
		startTime: time.Now(),
	}
}

func (pm *ProgressMonitor) OnProgress(uploaded, total int64) {
	now := time.Now()

	// 计算进度百分比
	percentage := float64(uploaded) / float64(total) * 100

	// 计算上传速度
	var speed float64
	if !pm.lastTime.IsZero() {
		duration := now.Sub(pm.lastTime).Seconds()
		if duration > 0 {
			bytesPerSecond := float64(uploaded-pm.lastUploaded) / duration
			speed = bytesPerSecond / 1024 // KB/s
		}
	}

	// 计算剩余时间
	var eta string
	if speed > 0 && uploaded > 0 {
		remainingBytes := total - uploaded
		remainingSeconds := float64(remainingBytes) / (speed * 1024)
		eta = FormatDuration(time.Duration(remainingSeconds) * time.Second)
	} else {
		eta = "Calculating..."
	}
	if pm.callback != nil {
		pm.callback(uploaded, total, percentage, speed, eta)
	} else {
		func(uploaded, total int64, percentage float64, speed float64, eta string) {
			progressBar := CreateProgressBar(int(percentage), 50)
			uploadedStr := FormatBytes(uploaded)
			totalStr := FormatBytes(total)
			speedStr := fmt.Sprintf("%.1f KB/s", speed)
			fmt.Printf("\r%s %.1f%% (%s/%s) %s ETA: %s",
				progressBar, percentage, uploadedStr, totalStr, speedStr, eta)
		}(uploaded, total, percentage, speed, eta)
	}
	pm.lastTime = now
	pm.lastUploaded = uploaded
}

func (pm *ProgressMonitor) GetTotalDuration() time.Duration {
	return time.Since(pm.startTime)
}
