package main

import (
	"log"
	"net/http"
	"os"

	"github.com/vladoiliev02/online-store/controller"
	"github.com/vladoiliev02/online-store/controller/security"
	"github.com/vladoiliev02/online-store/dao"
	"github.com/vladoiliev02/online-store/frontend"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	port   string
	host   string
	router chi.Router
)

func init() {
	initDb()
	initServer()
}

func main() {
	log.Println("Welcome to the store")
	http.ListenAndServe(":"+port, router)
}

func initDb() {
	driverName := getEnvVar("DB_DRIVER_NAME")
	connectionString := getEnvVar("DB_CONNECTION_STRING")

	dbOptions := dao.DBOptions{
		DriverName: driverName,
		ConnStr:    connectionString,
	}

	dao.Init(&dbOptions)
}

func initServer() {
	var exists bool
	port, exists = os.LookupEnv("PORT")
	if !exists {
		port = "8080"
	}

	host = getEnvVar("HOST")
	sessionStoreKey := os.Getenv("SESSION_STORE_KEY")
	clientID := getEnvVar("CLIENT_ID")
	clientSecret := getEnvVar("CLIENT_SECRET")
	oauthConfig := &security.OAuthConfiguration{
		Config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  host + "/oauth/code",
			Scopes:       []string{"openid", "profile", "email"},
			Endpoint:     google.Endpoint,
		},
		UserEndpoint: "https://www.googleapis.com/oauth2/v3/userinfo",
		LogoutPath:   "/logout",
		HomePath:     "/store/",
	}
	securityConfig := security.NewSecurityConfiguration(router, oauthConfig, sessionStoreKey)

	router = chi.NewMux()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	securityConfig.ConfigureRouter(router)

	frontend.Init(router)
	router.Mount("/api/v1", controller.Router())
}

func getEnvVar(name string) string {
	val, exists := os.LookupEnv(name)
	if !exists {
		log.Panic("Provide an environment variable: " + name)
	}

	return val
}
