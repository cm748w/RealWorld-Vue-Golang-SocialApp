package controllers

import (
	"Server/database"
	"Server/models"
	"context"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}

	return parsed
}

// CreatePost 创建帖子
// @Summary 创建帖子
// @Description 创建一条新帖子
// @Tags Posts
// @Accept json
// @Produce json
// @Param post body models.CreateOrUpdatePost true "创建帖子参数"
// @Success 200 {object} models.PostModel
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /posts [post]
func CreatePost(c *fiber.Ctx) error {

	var UserSchema = database.DB.Collection("users")
	var PostSchema = database.DB.Collection("posts")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var body models.CreateOrUpdatePost
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// 组装帖子数据
	var post models.PostModel
	post.Creator = c.Locals("userId").(string)
	post.Likes = make([]string, 0)
	post.Comments = make([]string, 0)
	post.CreatedAt = time.Now()
	post.Title = body.Title
	post.Message = body.Message
	post.SelectedFile = body.SelectedFile
	//

	var user models.UserModel
	objId, err := primitive.ObjectIDFromHex(c.Locals("userId").(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id",
		})
	}
	err = UserSchema.FindOne(ctx, bson.M{"_id": objId}).Decode(&user)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	//
	post.Name = user.Name
	// 数据组装完成
	// 创建帖子
	result, err := PostSchema.InsertOne(ctx, &post)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	} else {
		var createdPost *models.PostModel
		query := bson.M{"_id": result.InsertedID}

		if err := PostSchema.FindOne(ctx, query).Decode(&createdPost); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to load created post",
			})
		}
		return c.Status(fiber.StatusCreated).JSON(createdPost)
	}

}

// GetPost 获取帖子
// @Summary 获取单个帖子
// @Description 根据帖子 ID 获取帖子详情
// @Tags Posts
// @Accept json
// @Produce json
// @Param id path string true "帖子 ID"
// @Success 200 {object} models.PostModel
// @Failure 400 {object} map[string]interface{}
// @Router /posts/{id} [get]
func GetPost(c *fiber.Ctx) error {

	var PostSchema = database.DB.Collection("posts")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "post id is required",
		})
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	var post *models.PostModel
	query := bson.M{"_id": objID}

	err = PostSchema.FindOne(ctx, query).Decode(&post)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "post Not Found",
				"error":   err.Error(),
			})
		}

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "post Not Found",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(
		fiber.Map{
			"post": post,
		})

}

// UpdatePost 更新帖子
// @Summary 更新帖子
// @Description 更新指定 ID 的帖子
// @Tags Posts
// @Accept json
// @Produce json
// @Param id path string true "帖子 ID"
// @Param post body models.CreateOrUpdatePost true "更新帖子参数"
// @Success 200 {object} models.PostModel
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /posts/{id} [patch]
func UpdatePost(c *fiber.Ctx) error {

	var PostSchema = database.DB.Collection("posts")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var newData models.CreateOrUpdatePost
	if err := c.BodyParser(&newData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}
	// 权限校验开始
	var authPost models.PostModel
	primID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if err := PostSchema.FindOne(ctx, bson.M{"_id": primID}).Decode(&authPost); err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "post not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to load post",
		})
	}

	if authPost.Creator != c.Locals("userId").(string) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You Are Not authorized to update this post.",
		})
	}

	// 更新帖子字段
	authPost.Title = newData.Title
	authPost.Message = newData.Message
	authPost.SelectedFile = newData.SelectedFile
	// 执行更新
	_, err = PostSchema.UpdateOne(ctx, bson.M{"_id": authPost.ID}, bson.M{"$set": authPost})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{"data": err.Error()})
	} else {
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"post": authPost})
	}

}

// GetAllPosts 获取帖子列表
// @Summary 获取帖子列表
// @Description 分页获取帖子列表
// @Tags Posts
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param limit query int false "每页数量"
// @Param id query string true "用户 ID"
// @Success 200 {object} []models.PostModel
// @Failure 400 {object} map[string]interface{}
// @Router /posts [get]
func GetAllPosts(c *fiber.Ctx) error {

	var PostSchema = database.DB.Collection("posts")
	var userSchema = database.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var user models.UserModel
	var posts []models.PostModel

	userid := c.Query("id")
	if userid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user id is required",
		})
	}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "page must be a positive integer",
		})
	}

	defaultLimit := getEnvInt("POSTS_PAGE_SIZE", 2)
	maxLimit := getEnvInt("POSTS_MAX_PAGE_SIZE", 50)
	limit, err := strconv.Atoi(c.Query("limit", strconv.Itoa(defaultLimit)))
	if err != nil || limit < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "limit must be a positive integer",
		})
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	// 获取当前用户关注列表，并补上当前用户自身 ID
	mainUserID, err := primitive.ObjectIDFromHex(userid)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id",
		})
	}

	if err := userSchema.FindOne(ctx, bson.M{"_id": mainUserID}).Decode(&user); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	following := make([]string, 0, len(user.Following)+1)
	seen := make(map[string]struct{}, len(user.Following)+1)
	for _, followingID := range user.Following {
		if _, exists := seen[followingID]; exists {
			continue
		}
		seen[followingID] = struct{}{}
		following = append(following, followingID)
	}
	if _, exists := seen[userid]; !exists {
		following = append(following, userid)
	}
	///

	findOptions := options.Find()
	filter := bson.M{"creator": bson.M{"$in": following}}

	// 统计帖子总数
	total, err := PostSchema.CountDocuments(ctx, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to count posts",
		})
	}

	findOptions.SetSkip(int64((page - 1) * limit))
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.D{{Key: "_id", Value: -1}})

	// 查询帖子数据
	cursor, err := PostSchema.Find(ctx, filter, findOptions)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to query posts",
		})
	}

	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var post models.PostModel
		if err := cursor.Decode(&post); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to decode post",
			})
		}
		posts = append(posts, post)
	}

	if err := cursor.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "cursor iteration failed",
		})
	}

	if posts == nil {
		posts = make([]models.PostModel, 0)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":          posts,
		"currentPage":   page,
		"numberOfPages": (total + int64(limit) - 1) / int64(limit),
	})
}

// GetPostsUsersBySearch 搜索帖子与用户
// @Summary 按关键字搜索帖子与用户
// @Description 根据搜索关键字匹配帖子和用户
// @Tags Posts
// @Accept json
// @Produce json
// @Param searchQuery query string true "搜索关键词"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /posts/search [get]
func GetPostsUsersBySearch(c *fiber.Ctx) error {

	var PostSchema = database.DB.Collection("posts")
	var userSchema = database.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var users []models.UserModel
	var posts []models.PostModel

	// 初始化查询条件
	filterPost := bson.M{}
	filterUser := bson.M{}

	// 初始化查询选项
	findOptionsPost := options.Find()
	findOptionsUser := options.Find()

	if search := c.Query("searchQuery"); search != "" {
		// 帖子条件
		filterPost = bson.M{
			"$or": []bson.M{
				{
					"title": bson.M{
						"$regex": primitive.Regex{
							Pattern: search,
							Options: "i",
						},
					},
				},
				{
					"message": bson.M{
						"$regex": primitive.Regex{
							Pattern: search,
							Options: "i",
						},
					},
				},
			},
		}
		// 用户条件
		filterUser = bson.M{
			"$or": []bson.M{
				{
					"name": bson.M{
						"$regex": primitive.Regex{
							Pattern: search,
							Options: "i",
						},
					},
				},
				{
					"email": bson.M{
						"$regex": primitive.Regex{
							Pattern: search,
							Options: "i",
						},
					},
				},
			},
		}
	}
	// 执行查询
	cursorPosts, _ := PostSchema.Find(ctx, filterPost, findOptionsPost)
	defer cursorPosts.Close(ctx)

	cursorUsers, _ := userSchema.Find(ctx, filterUser, findOptionsUser)
	defer cursorUsers.Close(ctx)
	// 汇总结果

	for cursorUsers.Next(ctx) {
		var user models.UserModel
		cursorUsers.Decode(&user)
		users = append(users, user)
	}

	for cursorPosts.Next(ctx) {
		var post models.PostModel
		cursorPosts.Decode(&post)
		posts = append(posts, post)

	}

	return c.JSON(fiber.Map{
		"user":  users,
		"posts": posts,
	})
}

// CommentPost 评论帖子
// @Summary 评论帖子
// @Description 对指定帖子新增评论
// @Tags Posts
// @Accept json
// @Produce json
// @Param id path string true "帖子 ID"
// @Param post body models.CommentPost true "评论内容"
// @Success 200 {object} models.PostModel
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /posts/{id}/commentPost [post]
func CommentPost(c *fiber.Ctx) error {

	var PostSchema = database.DB.Collection("posts")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var b models.CommentPost
	if err := c.BodyParser(&b); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	postid, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"details": err.Error(),
		})
	}

	var post models.PostModel
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err = PostSchema.FindOneAndUpdate(ctx,
		bson.M{"_id": postid},
		bson.M{"$push": bson.M{"comments": b.Value}},
		opts,
	).Decode(&post)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"details": err.Error(),
		})
	}
	// TODO: 创建评论通知
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": post,
	})
}

// LikePost 点赞/取消点赞
// @Summary 点赞/取消点赞帖子
// @Description 如果当前用户未点赞则添加，已点赞则移除。
// @Tags Posts
// @Accept json
// @Produce json
// @Param id path string true "帖子 ObjectID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /posts/{id}/likePost [patch]
func LikePost(c *fiber.Ctx) error {

	var PostSchema = database.DB.Collection("posts")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	postid, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"details": "invalid post id",
		})
	}

	userID, errb := c.Locals("userId").(string)
	if !errb {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"details": "You are not authorized",
		})
	}

	var post models.PostModel
	pipeline := mongo.Pipeline{
		{{Key: "$set", Value: bson.M{
			"likes": bson.M{
				"$cond": bson.A{
					bson.M{"$in": bson.A{userID, "$likes"}},
					bson.M{
						"$filter": bson.M{
							"input": "$likes",
							"as":    "likeUserId",
							"cond":  bson.M{"$ne": bson.A{"$$likeUserId", userID}},
						},
					},
					bson.M{"$concatArrays": bson.A{"$likes", bson.A{userID}}},
				},
			},
		}}},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err = PostSchema.FindOneAndUpdate(ctx, bson.M{"_id": postid}, pipeline, opts).Decode(&post)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"details": "post not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"post": post,
	})

}

// DeletePost 删除帖子
// @Summary 根据 ID 删除帖子
// @Description 按帖子 ID 删除帖子，需提供创建者的认证 Token
// @Tags Posts
// @Accept json
// @Produce json
// @Param id path string true "帖子 ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /posts/{id} [delete]
func DeletePost(c *fiber.Ctx) error {

	var PostSchema = database.DB.Collection("posts")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var authPost models.PostModel
	primID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	err = PostSchema.FindOne(ctx, bson.M{"_id": primID}).Decode(&authPost)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "post not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"details": err.Error(),
		})
	}

	if authPost.Creator != c.Locals("userId").(string) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You are not authorized to delete this post",
		})
	}

	result, err := PostSchema.DeleteOne(ctx, bson.M{"_id": primID})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"details": err.Error(),
		})
	}

	if result.DeletedCount == 1 {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "post deleted successfully",
		})
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"message": "post not found",
	})

}
