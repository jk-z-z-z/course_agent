package bootstrap

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"course_agent_backend/internal/config"
	"course_agent_backend/internal/handler"
	"course_agent_backend/internal/middleware"
	"course_agent_backend/internal/model"
	"course_agent_backend/internal/repository"
	"course_agent_backend/internal/router"
	"course_agent_backend/internal/service"
)

type App struct {
	Config *config.Config
	MySQL  *MySQLClient
	Redis  *RedisClient
	Engine *gin.Engine
	Server *http.Server
}

type MySQLClient struct {
	DB *gorm.DB
}

type RedisClient struct {
	Client *redis.Client
}

func New(cfg *config.Config) (*App, error) {
	mysqlClient, err := newMySQLClient(cfg.MySQL)
	if err != nil {
		return nil, fmt.Errorf("init mysql: %w", err)
	}
	if err := mysqlClient.Ping(); err != nil {
		_ = mysqlClient.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	redisClient, err := newRedisClient(cfg.Redis)
	if err != nil {
		_ = mysqlClient.Close()
		return nil, fmt.Errorf("init redis: %w", err)
	}
	if err := redisClient.Ping(); err != nil {
		_ = redisClient.Close()
		_ = mysqlClient.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	if err := mysqlClient.DB.AutoMigrate(
		&model.User{},
		&model.Course{},
		&model.CourseMember{},
		&model.CourseStorageSpace{},
		&model.CourseMaterialNode{},
		&model.CourseMaterialVersion{},
		&model.CourseAgent{},
		&model.AgentConversation{},
		&model.AgentMessage{},
		&model.AgentMessageSource{},
	); err != nil {
		_ = redisClient.Close()
		_ = mysqlClient.Close()
		return nil, fmt.Errorf("migrate mysql: %w", err)
	}

	userRepo := repository.NewUserRepository(mysqlClient.DB)
	userService := service.NewUserService(userRepo, redisClient.Client)
	userHandler := handler.NewUserHandler(userService)
	authMiddleware := middleware.NewAuthMiddleware(userService)

	courseRepo := repository.NewCourseRepository(mysqlClient.DB)
	materialRepo := repository.NewMaterialRepository(mysqlClient.DB)
	agentRepo := repository.NewAgentRepository(mysqlClient.DB)
	courseService := service.NewCourseService(
		courseRepo,
		userRepo,
		materialRepo,
		agentRepo,
		cfg.Storage.RootPath,
		cfg.Storage.QuotaBytes,
		cfg.Agent.DefaultAgentName,
		cfg.Agent.PromptTemplate,
	)
	courseHandler := handler.NewCourseHandler(courseService)
	materialService := service.NewMaterialService(courseRepo, materialRepo)
	materialHandler := handler.NewMaterialHandler(materialService)

	engine := router.New(userHandler, courseHandler, materialHandler, authMiddleware, cfg.Server.Mode)
	addr := net.JoinHostPort(cfg.Server.Host, fmt.Sprintf("%d", cfg.Server.Port))
	server := &http.Server{
		Addr:              addr,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{
		Config: cfg,
		MySQL:  mysqlClient,
		Redis:  redisClient,
		Engine: engine,
		Server: server,
	}, nil
}

func newMySQLClient(cfg config.MySQLConfig) (*MySQLClient, error) {
	dsn, err := buildMySQLDSN(cfg)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}

	sqlDB.SetMaxOpenConns(config.DefaultMySQLMaxOpenConns())
	sqlDB.SetMaxIdleConns(config.DefaultMySQLMaxIdleConns())
	sqlDB.SetConnMaxLifetime(config.DefaultMySQLConnMaxLifetime())

	return &MySQLClient{DB: db}, nil
}

func (c *MySQLClient) Ping() error {
	if c == nil || c.DB == nil {
		return nil
	}
	sqlDB, err := c.DB.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return sqlDB.PingContext(ctx)
}

func (c *MySQLClient) Close() error {
	if c == nil || c.DB == nil {
		return nil
	}
	sqlDB, err := c.DB.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}
	return sqlDB.Close()
}

func newRedisClient(cfg config.RedisConfig) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:            cfg.Addr,
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        config.DefaultRedisPoolSize(),
		MinIdleConns:    config.DefaultRedisMinIdleConns(),
		DialTimeout:     config.DefaultRedisDialTimeout(),
		ReadTimeout:     config.DefaultRedisReadTimeout(),
		WriteTimeout:    config.DefaultRedisWriteTimeout(),
		ConnMaxIdleTime: config.DefaultRedisConnMaxIdleTime(),
	})
	return &RedisClient{Client: client}, nil
}

func (c *RedisClient) Ping() error {
	if c == nil || c.Client == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}
	return nil
}

func (c *RedisClient) Close() error {
	if c == nil || c.Client == nil {
		return nil
	}
	return c.Client.Close()
}

func buildMySQLDSN(cfg config.MySQLConfig) (string, error) {
	locName := config.DefaultMySQLLoc()
	loc, err := time.LoadLocation(locName)
	if err != nil {
		return "", fmt.Errorf("load mysql loc: %w", err)
	}

	params := url.Values{}
	params.Set("charset", config.DefaultMySQLCharset())
	if config.DefaultMySQLParseTime() {
		params.Set("parseTime", "true")
	} else {
		params.Set("parseTime", "false")
	}
	params.Set("loc", loc.String())

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		params.Encode(),
	)
	return dsn, nil
}

func (a *App) Close() error {
	if a == nil {
		return nil
	}
	var firstErr error
	if a.Redis != nil {
		if err := a.Redis.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if a.MySQL != nil {
		if err := a.MySQL.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
