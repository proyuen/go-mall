package utils

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRandomInt uses Table-Driven tests to cover edge cases.
func TestRandomInt(t *testing.T) {
	t.Parallel() // 允许并行测试，加快 CI 速度

	tests := []struct {
		name string
		min  int
		max  int
	}{
		{"Normal Range", 10, 20},
		{"Zero Range", 0, 10},
		{"Single Value (Min=Max)", 5, 5},
		{"Negative Range", -10, -1},
		{"Inverted Range (Min > Max)", 10, 5}, // 根据你的代码逻辑，这应该返回 Min
	}

	for _, tt := range tests {
		tt := tt // 捕获闭包变量
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// 运行 100 次以增加置信度
			for i := 0; i < 100; i++ {
				got := RandomInt(tt.min, tt.max)

				if tt.min >= tt.max {
					// 你的逻辑是 min >= max 直接返回 min
					assert.Equal(t, tt.min, got, "Should return min when min >= max")
				} else {
					assert.GreaterOrEqual(t, got, tt.min)
					assert.LessOrEqual(t, got, tt.max)
				}
			}
		})
	}
}

// TestRandomString verifies length and character set.
func TestRandomString(t *testing.T) {
	t.Parallel()

	// 预编译正则，不要在循环里 compile
	pattern := regexp.MustCompile(`^[a-z]+$`)
	lengths := []int{1, 10, 50, 100}

	for _, n := range lengths {
		str := RandomString(n)

		assert.Equal(t, n, len(str), "Length mismatch")
		assert.True(t, pattern.MatchString(str), "String contains invalid characters")
	}

	// Edge case: length 0
	assert.Empty(t, RandomString(0), "Should be empty for n=0")
}

// TestRandomOwner handles the variable length logic correctly.
func TestRandomOwner(t *testing.T) {
	t.Parallel()

	for i := 0; i < 50; i++ {
		owner := RandomOwner()

		// 你的逻辑是 RandomInt(3, 10)，所以长度必须在 [3, 10] 之间
		assert.GreaterOrEqual(t, len(owner), 3, "Owner name too short")
		assert.LessOrEqual(t, len(owner), 10, "Owner name too long")

		// 简单的字符集检查
		for _, r := range owner {
			assert.GreaterOrEqual(t, r, 'a')
			assert.LessOrEqual(t, r, 'z')
		}
	}
}

// TestRandomEmail covers default and custom domains.
func TestRandomEmail(t *testing.T) {
	t.Parallel()

	// 修正后的正则：匹配 3-10 位的用户名
	defaultPattern := regexp.MustCompile(`^[a-z]{3,10}@test\.com$`)

	t.Run("Default Domain", func(t *testing.T) {
		for i := 0; i < 50; i++ {
			email := RandomEmail("")
			assert.True(t, defaultPattern.MatchString(email), "Format error: "+email)
		}
	})

	t.Run("Custom Domain", func(t *testing.T) {
		domain := "google.com"
		customPattern := regexp.MustCompile(`^[a-z]{3,10}@google\.com$`)

		for i := 0; i < 50; i++ {
			email := RandomEmail(domain)
			assert.True(t, customPattern.MatchString(email), "Format error with custom domain: "+email)
		}
	})
}

// FuzzRandomInt uses Go 1.18+ Fuzzing to find crashes.
// 运行命令: go test -fuzz=Fuzz -fuzztime=10s
func FuzzRandomInt(f *testing.F) {
	// Seed corpus (初始语料)
	f.Add(0, 10)
	f.Add(10, 20)
	f.Add(-100, 100)

	f.Fuzz(func(t *testing.T, min int, max int) {
		// 这里不需要 mock，直接调用真实逻辑
		val := RandomInt(min, max)

		if min >= max {
			assert.Equal(t, min, val)
		} else {
			assert.GreaterOrEqual(t, val, min)
			assert.LessOrEqual(t, val, max)
		}
	})
}
