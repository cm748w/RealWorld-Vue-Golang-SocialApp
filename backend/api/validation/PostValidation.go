package validation

import (
	"Server/models"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// ValidatorPost 全局帖子验证器实例
var ValidatorPost = validator.New()

// ValidatePost 验证帖子数据
// @Summary 验证帖子数据
// @Description 验证帖子数据是否符合要求
// @Tags Validation
// @Accept json
// @Produce json
// @Param post body models.PostModel true "帖子数据"
// @Success 200 {string} string "验证通过"
// @Failure 400 {array} models.IError "验证失败"
// @Router /validate/post [post]
func ValidatePost(c *fiber.Ctx) error {
	var errors []*models.IError
	var body models.PostModel

	// 解析请求体
	if err := c.BodyParser(&body); err != nil {
		return err
	}

	// 验证结构体
	err := ValidatorPost.Struct(body)
	if err != nil {
		// 收集验证错误
		for _, err := range err.(validator.ValidationErrors) {
			var el models.IError
			el.Field = err.Field() // 错误字段
			el.Tag = err.Tag()     // 错误标签
			errors = append(errors, &el)
		}
		// 返回验证错误
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}
	// 校验通过，继续下一个中间件
	return c.Next()
}
