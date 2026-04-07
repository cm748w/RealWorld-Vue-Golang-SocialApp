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

// Create Post
// @Summary create a new post
// @Description create new post
// @Tags Posts
// @Accept json
// @Produce json
// @Param post body models.CreateOrUpdatePost true "post create details"
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

	// start set data
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
	// set data end
	// create post
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

// Get Post
// @Summary Get  a new post
// @Description Get a new post
// @Tags Posts
// @Accept json
// @Produce json
// @Param id path string true "Post id"
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

// Update Post
// @Summary Update post
// @Description Update post
// @Tags Posts
// @Accept json
// @Produce json
// @Param id path string true "Post id"
// @Param post body models.CreateOrUpdatePost true "update post details"
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
	// authorization start
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

	// set data end
	authPost.Title = newData.Title
	authPost.Message = newData.Message
	authPost.SelectedFile = newData.SelectedFile
	// create post
	_, err = PostSchema.UpdateOne(ctx, bson.M{"_id": authPost.ID}, bson.M{"$set": authPost})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{"data": err.Error()})
	} else {
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"post": authPost})
	}

}

// GetAllPosts
// @Summary get all posts
// @Description get all posts with pagination
// @Tags Posts
// @Accept json
// @Produce json
// @Param page query int false "page number"
// @Param limit query int false "page size"
// @Param id query string true "user id"
// @Success 200 {object} []models.PostModel
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
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

	// get user following list ides and add our user id to it
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

	// get total num of posts
	total, err := PostSchema.CountDocuments(ctx, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to count posts",
		})
	}

	findOptions.SetSkip(int64((page - 1) * limit))
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.D{{Key: "_id", Value: -1}})

	// start get posts
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

// GetPostsUsersBySearch
// @Summary get posts users by search
// @Description get posts users maching the search query
// @Tags Posts
// @Accept json
// @Produce json
// @Param searchQuery query string true "search query"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /posts/search [get]
func GetPostsUsersBySearch(c *fiber.Ctx) error {

	var PostSchema = database.DB.Collection("posts")
	var userSchema = database.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var users []models.UserModel
	var posts []models.PostModel

	//
	filterPost := bson.M{}
	filterUser := bson.M{}

	//
	findOptionsPost := options.Find()
	findOptionsUser := options.Find()

	if search := c.Query("searchQuery"); search != "" {
		// post
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
		//
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
	// end
	cursorPosts, _ := PostSchema.Find(ctx, filterPost, findOptionsPost)
	defer cursorPosts.Close(ctx)

	cursorUsers, _ := userSchema.Find(ctx, filterUser, findOptionsUser)
	defer cursorUsers.Close(ctx)
	//

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

// Comment Post
// @Summary comment post
// @Description comment post
// @Tags Posts
// @Accept json
// @Produce json
// @Param id path string true "Post id"
// @Param post body models.CommentPost true "comment value"
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
	// TODO create notification start
	// end
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": post,
	})
}
