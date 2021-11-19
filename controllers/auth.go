package controllers

import (
	"context"
	"fmt"
	"go-api/database"
	"go-api/models"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)


func Hello(c *fiber.Ctx) error {

	return c.JSON(fiber.Map{"message": "welcome"})
}

func Register(c *fiber.Ctx) error {

	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(data["password"]), 14)

	user := models.User{ID: primitive.NewObjectID(), Name: data["name"], Email: data["email"],
		Password: password}

	filter := bson.D{primitive.E{Key: "email", Value: data["email"]}}

	err := database.DB.FindOne(context.TODO(), filter).Decode(&user)

	if err == mongo.ErrNoDocuments {
		insertResult, err := database.DB.InsertOne(context.TODO(), user)
		if err != nil {
			return err
		}

		fmt.Println("Inserted a new user: ", insertResult.InsertedID)

		return c.JSON(user)
	}

	c.Status(fiber.StatusBadRequest)

	return c.JSON(fiber.Map{"message": "Email already exists"})
}

func Login(c *fiber.Ctx) error {

	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	var user models.User

	filter := bson.D{primitive.E{Key: "email", Value: data["email"]}}

	err := database.DB.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		c.Status(fiber.StatusNotFound)
		return c.JSON(fiber.Map{"message": "User not found"})
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(data["password"])); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{"message": "Wrong credentials"})
	}
	stringObjectID := (user.ID).Hex()

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    stringObjectID,
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	})

	token, err := claims.SignedString([]byte(os.Getenv("JWT_SECRET")))

	if err != nil {

		c.Status(fiber.StatusInternalServerError)

		return c.JSON(fiber.Map{"message": "Could not login"})
	}
	cookie := fiber.Cookie{

		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
	}

	c.Cookie(&cookie)

	return c.JSON(fiber.Map{"message": "Success"})
}

func User(c *fiber.Ctx) error {

	cookie := c.Cookies("jwt")

	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {

		c.Status(fiber.StatusUnauthorized)

		return c.JSON(fiber.Map{"message": "unauthorized"})
	}

	claims := token.Claims.(*jwt.StandardClaims)

	objID, _ := primitive.ObjectIDFromHex(claims.Issuer)

	var user models.User

	filter := bson.D{primitive.E{Key: "_id", Value: objID}}

	err1 := database.DB.FindOne(context.TODO(), filter).Decode(&user)

	if err1 != nil {
		c.Status(fiber.StatusNotFound)
		return c.JSON(fiber.Map{"message": "unauthorized token"})
	}

	return c.JSON(user)
}

func Logout(c *fiber.Ctx) error {

	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
	}

	c.Cookie(&cookie)

	return c.JSON(fiber.Map{"message": "Logout successful"})
}
