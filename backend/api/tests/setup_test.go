package tests

import (
	"Server/database"
	"Server/routes"
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var app *fiber.App

func TestMain(m *testing.M) {
	// Setup
	setup()

	// Run Tests
	code := m.Run()

	// cleanup
	cleanup()

	os.Exit(code)
}

func setup() {
	// load test env vars
	if err := godotenv.Load("../.env.test"); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			log.Printf("warning: could not load .env file: %v", err)
		}
	}

	// Set jwt if not provided
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "test-jwt-secret-key")
	}

	// Connect to test db
	connectTestDB()

	// setup fiber app
	app = fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowOriginsFunc: func(origin string) bool {
			return true
		},
	}))

	// setup routes
	routes.SetupAuthRoutes(app)
	//..
}

func connectTestDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// use test db
	mongoUri := os.Getenv("TEST_MONGO_URI")
	if mongoUri == "" {
		mongoUri = os.Getenv("MONGO_URI")
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))
	if err != nil {
		log.Fatal("Failed to connect to test db:", err)
	}
	database.Client = client
	database.DB = client.Database("social_test")

	// 与生产保持一致：唯一索引必须在测试库里也存在，否则「并发注册同邮箱被
	// 唯一索引兜底」这条分支永远测不到（历史上测试库因此与生产结构不一致）。
	if err := database.EnsureIndexes(ctx, database.DB); err != nil {
		log.Fatal("Failed to create indexes on test db:", err)
	}
}

func cleanup() {
	if database.Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// drop test db
		if err := database.DB.Drop(ctx); err != nil {
			log.Printf("cleanup: drop test db: %v", err)
		}
		if err := database.Client.Disconnect(ctx); err != nil {
			log.Printf("cleanup: disconnect: %v", err)
		}
	}
}

// helper for cleanup connections
func cleanupCollections() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collections := []string{"users"}
	for _, collection := range collections {
		if err := database.DB.Collection(collection).Drop(ctx); err != nil {
			log.Printf("cleanup: drop %s: %v", collection, err)
		}
	}
}
