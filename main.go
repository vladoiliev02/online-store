package main

import (
	"log"
	"net/http"
	"online-store/controller"
	"online-store/controller/security"
	"online-store/dao"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	port   string
	router chi.Router
)

func init() {
	initDb()
	initServer()
}

// PGPASSWORD=Test1234 psql -p 5342 -U postgres -h localhost -d online_store -f ./sql/createDatabase.sql
// go build . && ./online-store
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
	port = getEnvVar("PORT")
	sessionStoreKey := os.Getenv("SESSION_STORE_KEY")
	clientID := getEnvVar("CLIENT_ID")
	clientSecret := getEnvVar("CLIENT_SECRET")
	oauthConfig := &security.OAuthConfiguration{
		Config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  "http://localhost:30080/oauth/code",
			Scopes:       []string{"openid", "profile", "email"},
			Endpoint:     google.Endpoint,
		},
		UserEndpoint: "https://www.googleapis.com/oauth2/v3/userinfo",
	}
	securityConfig := security.NewSecurityConfiguration(router, oauthConfig, sessionStoreKey)
	fs := http.FileServer(http.Dir("./static"))

	router = chi.NewMux()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	securityConfig.ConfigureRouter(router)
	router.Mount("/static", http.StripPrefix("/static/", fs))
	router.Mount("/api/v1", controller.Router())
}

func getEnvVar(name string) string {
	val, exists := os.LookupEnv(name)
	if !exists {
		log.Panic("Provide an environment variable: " + name)
	}

	return val
}
