package service

import (
	"testing"

	"golang-learn/internal/model"
)

// TestFavorToTemperature 好感度 → 关系温度映射（表驱动测试）
// 覆盖：五个档位的典型值 + 边界值（阈值本身）+ 非法值
func TestFavorToTemperature(t *testing.T) {
	tests := []struct {
		name  string               // 用例名
		favor int                  // 输入好感度
		want  model.Temperature    // 期望温度
	}{
		{"低于0拉档", -1, model.TempCold},
		{"0分拉档", 0, model.TempCold},
		{"39分拉档", 39, model.TempCold},
		{"40分凉档边界", 40, model.TempCool},
		{"59分凉档", 59, model.TempCool},
		{"60分温档边界", 60, model.TempMild},
		{"74分温档", 74, model.TempMild},
		{"75分热档边界", 75, model.TempWarm},
		{"89分热档", 89, model.TempWarm},
		{"90分夯档边界", 90, model.TempHot},
		{"100分满分", 100, model.TempHot},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FavorToTemperature(tt.favor)
			if got != tt.want {
				t.Errorf("FavorToTemperature(%d) = %v, 期望 %v",
					tt.favor, got, tt.want)
			}
		})
	}
}

// TestTemperatureLabel 温度中文标签
func TestTemperatureLabel(t *testing.T) {
	tests := []struct {
		temp   model.Temperature
		want   string
	}{
		{model.TempHot, "🔥 夯"},
		{model.TempWarm, "😊 热"},
		{model.TempMild, "🙂 温"},
		{model.TempCool, "😐 凉"},
		{model.TempCold, "🥶 拉"},
		{model.Temperature(99), "未知"},
	}
	for _, tt := range tests {
		if got := tt.temp.Label(); got != tt.want {
			t.Errorf("Temperature(%d).Label() = %q, 期望 %q", tt.temp, got, tt.want)
		}
	}
}
