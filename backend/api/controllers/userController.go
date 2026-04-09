package controllers

import (
	"Server/database"
	"Server/models"
	"context"
	"slices"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetUserByID 获取用户信息
// @Summary 按 ID 获取用户
// @Description 根据用户 ID 获取用户详情
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "用户 ID"
// @Success 201 {object} models.UserModel
// @Failure 400 {object} map[string]interface{}
// @Router /user/getUser/{id} [get]
func GetUserByID(c *fiber.Ctx) error {

	var UserSchema = database.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var user models.UserModel

	objId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid User id",
		})
	}
	// strID := c.Params("id")
	// TODO: 获取并返回该用户的帖子

	// 查询用户数据
	userResult := UserSchema.FindOne(ctx, bson.M{"_id": objId})

	if err := userResult.Err(); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "User Not found",
		})
	}

	if err := userResult.Decode(&user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "failed to load user",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user":  user,
		"posts": "posts",
	})
}

// UpdateUser 更新用户信息
// @Summary 更新用户资料
// @Description 更新当前登录用户资料
// @Tags Users
// @Accept json
// @Produce json
// @Param user body models.UpdateUser true "用户更新参数"
// @Success 201 {object} models.UserModel
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /user/update [patch]
func UpdateUser(c *fiber.Ctx) error {

	var UserSchema = database.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	extUid, ok := c.Locals("userId").(string)
	if !ok || extUid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized user",
		})
	}

	userid, err := primitive.ObjectIDFromHex(extUid)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid User id",
		})
	}

	var user models.UpdateUser
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	update := bson.M{"name": user.Name, "imageUrl": user.ImageUrl, "bio": user.Bio}

	result, err := UserSchema.UpdateOne(ctx, bson.M{"_id": userid}, bson.M{"$set": update})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "cannot update the user data",
			"details": err.Error(),
		})
	}
	//
	var updateUser models.UserModel // 原变量名写错，这里使用更清晰的命名
	if result.MatchedCount == 1 {
		err := UserSchema.FindOne(ctx, bson.M{"_id": userid}).Decode(&updateUser)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"details": err.Error(),
			})
		}
	} else {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": updateUser})

}

// FollowingUser 关注/取消关注用户
// @Summary 关注/取消关注用户
// @Description 对目标用户执行关注或取消关注
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "用户 ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /user/{id}/following [patch]
func FollowingUser(c *fiber.Ctx) error {

	var UserSchema = database.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var FirstUser models.UserModel
	var SecondUser models.UserModel

	FirstUserID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"details": "Invalid target user id",
		})
	}

	suid, ok := c.Locals("userId").(string)
	if !ok || suid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"details": "invalid auth user id",
		})
	}

	SecondUserID, err := primitive.ObjectIDFromHex(suid)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"details": "Invalid auth user id",
		})
	}

	if FirstUserID == SecondUserID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"details": "you cannot follow yourself",
		})
	}

	err = UserSchema.FindOne(ctx, bson.M{"_id": FirstUserID}).Decode(&FirstUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"details": "target user not found",
			})
		}

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"details": err.Error(),
		})
	}

	err = UserSchema.FindOne(ctx, bson.M{"_id": SecondUserID}).Decode(&SecondUser)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"details": "auth user not found",
			})
		}

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"details": err.Error(),
		})
	}

	fuid := c.Params("id")

	if slices.Contains(FirstUser.Followers, suid) {
		if i := slices.Index(FirstUser.Followers, suid); i >= 0 {
			FirstUser.Followers = slices.Delete(FirstUser.Followers, i, i+1)
		}
		// 从当前用户的 following 列表中移除
		if index := slices.Index(SecondUser.Following, fuid); index >= 0 {
			SecondUser.Following = slices.Delete(SecondUser.Following, index, index+1)
		}
	} else {
		FirstUser.Followers = append(FirstUser.Followers, suid)
		SecondUser.Following = append(SecondUser.Following, fuid)

		// TODO: 创建关注通知
	}

	updateFirst := bson.M{"followers": FirstUser.Followers}
	updateSecond := bson.M{"following": SecondUser.Following}

	_, err = UserSchema.UpdateOne(ctx, bson.M{"_id": FirstUserID}, bson.M{"$set": updateFirst})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"details": err.Error(),
		})
	}

	_, err = UserSchema.UpdateOne(ctx, bson.M{"_id": SecondUserID}, bson.M{"$set": updateSecond})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"details": err.Error(),
		})
	}

	err = UserSchema.FindOne(ctx, bson.M{"_id": FirstUserID}).Decode(&FirstUser)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"details": err.Error(),
		})
	}

	err = UserSchema.FindOne(ctx, bson.M{"_id": SecondUserID}).Decode(&SecondUser)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"details": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"SecondUser": SecondUser, "FirstUser": FirstUser})

}

// GetSugUser 获取推荐用户
// @Summary 获取推荐用户
// @Description 基于当前用户关注关系获取推荐用户
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /user/getSug [get]
func GetSugUser(c *fiber.Ctx) error {

	var UserSchema = database.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var MainUser models.UserModel
	var SugListId []primitive.ObjectID
	var AllSugUsers []models.UserModel
	seenSugIDs := make(map[primitive.ObjectID]struct{})
	alreadyFollowing := make(map[primitive.ObjectID]struct{})

	mainUserHex, ok := c.Locals("userId").(string)
	if !ok || mainUserHex == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"details": "invalid auth user id",
		})
	}

	MainUserID, err := primitive.ObjectIDFromHex(mainUserHex)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"details": err.Error(),
		})
	}

	err = UserSchema.FindOne(ctx, bson.M{"_id": MainUserID}).Decode(&MainUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"details": "main user not found",
			})
		}

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"details": err.Error(),
		})
	}

	for _, fidHex := range MainUser.Following {
		fid, err := primitive.ObjectIDFromHex(fidHex)
		if err != nil {
			continue
		}
		alreadyFollowing[fid] = struct{}{}
	}

	// 收集推荐用户 ID 列表
	for _, fidHex := range MainUser.Following {
		var singleUser models.UserModel

		convFID, err := primitive.ObjectIDFromHex(fidHex)
		if err != nil {
			continue
		}

		err = UserSchema.FindOne(ctx, bson.M{"_id": convFID}).Decode(&singleUser)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				continue
			}

			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"details": err.Error(),
			})
		}

		// 从对方的 following 中提取推荐用户
		for _, idHex := range singleUser.Following {
			convID, err := primitive.ObjectIDFromHex(idHex)
			if err != nil {
				continue
			}
			if convID != MainUserID {
				if _, followed := alreadyFollowing[convID]; followed {
					continue
				}
				if _, exists := seenSugIDs[convID]; !exists {
					seenSugIDs[convID] = struct{}{}
					SugListId = append(SugListId, convID)
				}
			}
		}

		// 从对方的 followers 中提取推荐用户
		for _, idHex := range singleUser.Followers {
			convID, err := primitive.ObjectIDFromHex(idHex)
			if err != nil {
				continue
			}
			if convID != MainUserID {
				if _, followed := alreadyFollowing[convID]; followed {
					continue
				}
				if _, exists := seenSugIDs[convID]; !exists {
					seenSugIDs[convID] = struct{}{}
					SugListId = append(SugListId, convID)
				}
			}
		}
	}

	// 根据推荐 ID 批量查询用户
	if len(SugListId) > 0 {

		// 使用 $in 一次性查询所有推荐用户
		cursor, err := UserSchema.Find(ctx, bson.M{
			"_id": bson.M{"$in": SugListId}, // 直接用 SugListId
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"details": err.Error(),
			})
		}

		defer cursor.Close(ctx)

		if err = cursor.All(ctx, &AllSugUsers); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"details": err.Error(),
			})
		}
	}

	if AllSugUsers == nil {
		AllSugUsers = make([]models.UserModel, 0)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"users": AllSugUsers})
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 删除当前登录用户
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /user/delete [delete]
func DeleteUser(c *fiber.Ctx) error {

	var UserSchema = database.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	extUid, ok := c.Locals("userId").(string)
	if !ok || extUid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized user",
		})
	}

	userID, err := primitive.ObjectIDFromHex(extUid)

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Invalid User id",
		})
	}

	result, err := UserSchema.DeleteOne(ctx, bson.M{"_id": userID})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "failed to delete user",
			"error":   err.Error(),
		})
	}

	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "user not found",
		})
	}
	// 删除成功
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User Deleted Successfully",
	})
}
