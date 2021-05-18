package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/VarthanV/pizza/pizza/services"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"

	"github.com/VarthanV/pizza/handlers"
	"github.com/VarthanV/pizza/middlewares"
	"github.com/VarthanV/pizza/pizza"
	pizzaImplementaion "github.com/VarthanV/pizza/pizza/implementation"
	pizzaRepo "github.com/VarthanV/pizza/pizza/repositories"
	"github.com/VarthanV/pizza/shared"
	"github.com/VarthanV/pizza/users"
	implementation "github.com/VarthanV/pizza/users/implementation"
	"github.com/VarthanV/pizza/users/repositories"
	"github.com/VarthanV/pizza/users/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang/glog"
)

func main() {
	var db *sql.DB
	{
		var err error
		err = godotenv.Load()
		if err != nil {
			glog.Fatalf("Unable to load environment variables")
		}
		// Initializing DB Constants
		dbConnection := shared.DBConnection{
			DBName:   os.Getenv("DB_NAME"),
			Password: os.Getenv("DB_PASSWORD"),
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			Username: os.Getenv("DB_USERNAME"),
		}
		// Initialize mysql database
		//<username>:<pw>@tcp(<HOST>:<port>)/<dbname>
		connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			dbConnection.Username,
			dbConnection.Password,
			dbConnection.Host,
			dbConnection.Port,
			dbConnection.DBName)
		db, err = sql.Open("mysql", connectionString)
		if err != nil {
			glog.Fatalf("Unable to connect to db...", err)
			os.Exit(-1)
		}
	}
	glog.Info("Connected to mysql db.....")
	var redisClient *redis.Client
	{
		redisClient = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		})
		// Make a ping
		ping := redisClient.Ping(context.TODO())
		result, _ := ping.Result()
		glog.Info("Result from redis ping...", result)
	}

	constants := shared.SharedConstants{
		AccessTokenSecretKey:  os.Getenv("ACCESS_TOKEN_SECRET_KEY"),
		RefreshTokenSecretKey: os.Getenv("REFRESH_TOKEN_SECRET_KEY"),
	}
	var utilityservice utils.UtilityService
	{
		utilityservice = utils.NewUtilityService(&constants)
	}
	var tokenService users.TokenService
	var usersvc users.Service

	{
		dbRepository, err := repositories.NewMySqlRepository(db)
		rdbRepository := repositories.NewRedisRepository(redisClient, utilityservice)
		tokenService = implementation.NewTokenService(rdbRepository)

		if err != nil {
			glog.Error("Unable to initialize dbRepository...", err)
			return
		}
		if rdbRepository != nil {
			glog.Error("Unable to initialize rdb Repository....")
		}
		glog.Info("Initializing user service....")
		usersvc = implementation.NewService(dbRepository, tokenService, utilityservice)
	}

	var pizzaService pizza.Service
	{
		pizzaRepo := pizzaRepo.NewPizzaMysqlRepository(db)
		pizzaService = pizzaImplementaion.NewService(pizzaRepo)
	}
	var cartService services.CartService
	{
		cartRepo := pizzaRepo.NewCartRepository(db)
		cartService = pizzaImplementaion.NewCartService(cartRepo)
	}
	glog.Info("Init handlers....")
	userHandler := handlers.NewUserHandler(usersvc)
	pizzaHandlers := handlers.NewPizzaHandler(pizzaService)
	cartHandlers := handlers.NewCartHandler(cartService)
	// Initializing middleware
	var middleware middlewares.Service
	{

		middleware = middlewares.NewMiddleware(&constants, tokenService)
	}
	router := gin.Default()

	//TODO: Mapping signup route with its respective handler.... Will be refactored later
	router.POST("/signup", userHandler.SignUpUserHandler)
	router.POST("/login", userHandler.LoginUserHandler)

	pizzaroutes := router.Group("/pizzas")
	{
		pizzaroutes.GET("", pizzaHandlers.GetAllPizzas)
	}
	// Group authRoute
	authenticated := router.Group("/cart", middleware.VerifyTokenMiddleware)
	{
		authenticated.GET("/", cartHandlers.GetCart)
		authenticated.POST("/add", cartHandlers.AddToCart)
	}

	// Run the router
	router.Run(":8080")

}
