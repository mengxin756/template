package repository

import (
	"testing"

	"example.com/classic/internal/domain"
	"github.com/stretchr/testify/assert"
)

// 使用真实的 Ent 客户端进行集成测试
func TestUserRepository_Integration(t *testing.T) {
	// 跳过集成测试，除非明确要求
	t.Skip("跳过集成测试，需要真实的数据库连接")

	// 这里可以添加真实的集成测试
}

// 简单的单元测试，测试领域逻辑
func TestUserRepository_DomainLogic(t *testing.T) {
	// 测试用户状态验证
	t.Run("用户状态验证", func(t *testing.T) {
		status := domain.Status("active")
		assert.True(t, status.IsValid())

		status = domain.Status("inactive")
		assert.True(t, status.IsValid())

		status = domain.Status("invalid")
		assert.False(t, status.IsValid())
	})

	// 测试用户查询构建
	t.Run("用户查询构建", func(t *testing.T) {
		status := domain.StatusActive
		query := &domain.UserQuery{
			Page:     1,
			PageSize: 10,
			Status:   &status,
		}

		assert.Equal(t, 1, query.Page)
		assert.Equal(t, 10, query.PageSize)
		assert.Equal(t, domain.StatusActive, *query.Status)
	})
}
