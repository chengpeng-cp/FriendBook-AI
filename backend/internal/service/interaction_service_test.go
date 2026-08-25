package service

import (
	"testing"
	"time"

	"golang-learn/internal/model"
)

// TestInteractDaysToTemperature 互动时间 → 关系温度推断（表驱动测试）
// 规则：<=7天夯 | <=30天热 | <=90天温 | <=180天凉 | >180天拉
func TestInteractDaysToTemperature(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		interactAgo time.Duration // 距今多久互动过
		want      model.Temperature
	}{
		{"今天刚互动", 1 * time.Hour, model.TempHot},
		{"7天内互动", 6 * 24 * time.Hour, model.TempHot},
		{"7天边界(第7天)", 7 * 24 * time.Hour, model.TempHot},
		{"30天内互动", 29 * 24 * time.Hour, model.TempWarm},
		{"30天边界", 30 * 24 * time.Hour, model.TempWarm},
		{"90天内互动", 89 * 24 * time.Hour, model.TempMild},
		{"90天边界", 90 * 24 * time.Hour, model.TempMild},
		{"180天内互动", 179 * 24 * time.Hour, model.TempCool},
		{"180天边界", 180 * 24 * time.Hour, model.TempCool},
		{"超过180天", 365 * 24 * time.Hour, model.TempCold},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InteractDaysToTemperature(now.Add(-tt.interactAgo))
			if got != tt.want {
				t.Errorf("距今 %v 互动，期望温度 %v，实际 %v",
					tt.interactAgo, tt.want, got)
			}
		})
	}
}
