package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/streadway/amqp"
	"github.com/superryanguo/kitchen/clients"
	"github.com/superryanguo/kitchen/cooks"
	implementation "github.com/superryanguo/kitchen/implementation"
	"github.com/superryanguo/kitchen/inmemorydb"
	"github.com/superryanguo/kitchen/migrations"
	"github.com/superryanguo/kitchen/mysql"
	"github.com/superryanguo/kitchen/processes"
	"github.com/superryanguo/kitchen/queue"
	"github.com/superryanguo/kitchen/redisclient"
	"github.com/superryanguo/kitchen/seeder"
	"github.com/superryanguo/kitchen/shared"
)

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func main() {
	prometheusClient := clients.NewPrometheusClient()
	var db *sql.DB
	{
		var err error
		//err = godotenv.Load()
		//if err != nil {
		//glog.Fatalf("Unable to load environment variables")
		//}
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
	// Msg queue

	rabbitMqConnection, err := amqp.Dial(os.Getenv("RABBIT_MQ_CONNECTION_STRING"))
	glog.Info("Kitchen Connection string is....", os.Getenv("RABBIT_MQ_CONNECTION_STRING"))
	if err != nil {
		glog.Fatalf("Unable to connect to rabbit mq %f", err)
	}

	// RMQ Channel
	ch, err := rabbitMqConnection.Channel()
	if err != nil {
		glog.Fatalf("Unable to create a channel %f", err)
	}

	migrationsvc := migrations.NewMigrationService(db)
	migrationsvc.RunMigrations(context.TODO())

	isSeedingEnabled := os.Getenv("SEEDING_ENABLED")
	// Seed data -> If not needed we can omit it env and write logic accordingly
	if isSeedingEnabled == "true" {
		seederSvc := seeder.NewSeederService(db)
		seederSvc.SeedData()
	}
	var redisClient *redis.Client
	{
		redisClient = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf(`%s:%s`, os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
			Password: "", // no password set
			DB:       0,  // use default DB
		})
		// Make a ping
		ping := redisClient.Ping(context.TODO())
		result, err := ping.Result()
		if err != nil {
			glog.Info("Error connecting to redis...", err)
		}
		glog.Info("Result from redis ping...", result)
	}

	var cookservice cooks.Service
	{
		cookRepo := mysql.NewCookMysqlRepo(db)
		cookservice = implementation.NewCookService(cookRepo)
	}

	var processUpdateService processes.OrderProcessUpdateService
	{
		repo := mysql.NewOrderProcessUpdateRepoMysql(db)
		processUpdateService = implementation.NewOrderOrderProcessUpdateImplementation(repo)
	}

	var orderRequestInmemoryService inmemorydb.OrderRequestInMemoryService
	{
		repo := redisclient.NewOrderQueueRepo(redisClient)
		orderRequestInmemoryService = implementation.NewOrderInmemoryService(repo)
	}
	queueRepo := queue.NewRabbitRepository(ch)

	var processOrderSvc processes.OrderProcessService
	{
		processOrderSvc = implementation.NewProcessOrderImplementationService(cookservice, processUpdateService, orderRequestInmemoryService, queueRepo, prometheusClient)
	}

	var orderRequestsvc processes.OrderRequestService
	{
		orderRequestsvc = implementation.NewOrderRequestImplementation(cookservice, processOrderSvc, orderRequestInmemoryService)
	}
	var queueService queue.QueueService
	{

		queueService = implementation.NewRabbitMQService(queueRepo, orderRequestsvc)
	}

	r := gin.Default()

	var responseStatus = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "response_status",
			Help: "Status of HTTP response",
		},
		[]string{"status"},
	)
	r.GET("/metrics", prometheusHandler())
	r.GET("/ping", func(c *gin.Context) {
		responseStatus.WithLabelValues(strconv.Itoa(200)).Inc()
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	ctx := context.Background()
	queueService.ConsumeOrderDetails(ctx)
	prometheus.MustRegister(responseStatus)
	prometheusClient.RegisterMetrics()
	r.Run(os.Getenv("PORT"))
}
