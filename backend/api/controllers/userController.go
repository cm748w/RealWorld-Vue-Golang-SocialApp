package controllers

import (
	"Server/database"
	"Server/models"
	"context"
	"slices"
	"sort"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetUserBy ID
// @Summary Get User By ID
// @Description GetUser Deatils By ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 201 {object} models.UserModel
// @Failure 400 {object} map[string]interface{}
// @Router /user/getUser/{id} [get]
func GetUserByID(c *fiber.Ctx) error {

	var UserSchema = database.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var user models.UserModel

	objId, _ := primitive.ObjectIDFromHex(c.Params("id"))
	// strID := c.Params("id")
	// TODO GET and REturn user posts

	// get nuser data
	userResult := UserSchema.FindOne(ctx, bson.M{"_id": objId})

	if userResult.Err() != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"success": false,
			"message": "User Not found",
		})
	}

	userResult.Decode(&user)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user":  user,
		"posts": "posts",
	})
}

// UpdateUser
// @Summary update user data
// @Description update user deatils
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body models.UpdateUser true "deatils "
// @Success 201 {object} models.UserModel
// @Failure 400 {object} map[string]interface{}
// @security BearerAuth
// @Router /user/Update/{id} [patch]
func UpdateUser(c *fiber.Ctx) error {

	var UserSchema = database.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//
	extUid := c.Locals("userId").(string)

	if extUid != c.Params("id") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "You Are Not Authroized to Update This Profile",
		})
	}

	userid, _ := primitive.ObjectIDFromHex(c.Params("id"))

	var user models.UpdateUser
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
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
	var updateUser models.UserModel // 老外写的是 updateUsser ,太抽象了 ，实在不能跟着写
	if result.MatchedCount == 1 {
		err := UserSchema.FindOne(ctx, bson.M{"_id": userid}).Decode(&updateUser)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"details": err.Error(),
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": updateUser})

}

// Following Users
// @Summary Follow/UnFollow User
// @Description follow or  un follow a user
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @security BearerAuth
// @Router /user/{id}/following [patch]
func FollowingUser(c *fiber.Ctx) error {

	var UserSchema = database.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var FirstUser models.UserModel
	var SecondUser models.UserModel

	FirstUserID, _ := primitive.ObjectIDFromHex(c.Params("id"))
	SecondUserID, _ := primitive.ObjectIDFromHex(c.Locals("userId").(string))

	err := UserSchema.FindOne(ctx, bson.M{"_id": FirstUserID}).Decode(&FirstUser)
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

	fuid := c.Params("id")
	suid := c.Locals("userId").(string)

	if slices.Contains(FirstUser.Followers, suid) {
		i := sort.SearchStrings(FirstUser.Followers, suid)
		FirstUser.Followers = slices.Delete(FirstUser.Followers, i, i+1)
		// remove from the following list on second user
		index := sort.SearchStrings(SecondUser.Following, fuid)
		SecondUser.Following = slices.Delete(SecondUser.Following, index, index+1)
	} else {
		FirstUser.Followers = append(FirstUser.Followers, suid)
		SecondUser.Following = append(SecondUser.Following, fuid)

		// TODO :: Create Notification
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

// GetSugUser Users
// @Summary Get Suggersted users
// @Description get suggested userses based on the current user's following list
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @security BearerAuth
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

	// Get SugUsers id then put them in suglistid
	for _, fidHex := range MainUser.Following {
		var singleUser models.UserModel

		convFID, err := primitive.ObjectIDFromHex(fidHex)
		if err != nil {
			continue
		}

		err = UserSchema.FindOne(ctx, bson.M{"_id": convFID}).Decode(&singleUser)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"details": err.Error(),
			})
		}

		// following
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

		// Followers
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

	// Gest Sug User by id .
	if len(SugListId) > 0 {

		// fetch all users in one qeery using $in operator
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
