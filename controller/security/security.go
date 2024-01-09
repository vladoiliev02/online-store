package security

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"online-store/controller"
	"online-store/dao"
	"online-store/model"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

const (
	sessionName     = "authentication"
	oauthStateKey   = "state"
	redirectBackKey = "redirectBack"
)

type OAuthConfiguration struct {
	oauth2.Config
	UserEndpoint string
	LogoutPath   string
	HomePath     string
}

type UserInfo struct {
	Sub        string `json:"sub"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	FirstName  string `json:"given_name"`
	LastName   string `json:"family_name"`
	PictureURL string `json:"picture"`
}

type SecurityConfiguration struct {
	store       *sessions.CookieStore
	oauthConfig *OAuthConfiguration
	userDAO     dao.UserDAO
}

func NewSecurityConfiguration(r chi.Router, oauthConfig *OAuthConfiguration, sessionStoreKey string) *SecurityConfiguration {
	return &SecurityConfiguration{
		store:       sessions.NewCookieStore([]byte(sessionStoreKey)),
		oauthConfig: oauthConfig,
		userDAO:     *dao.NewUserDAO(),
	}
}

func (sc *SecurityConfiguration) ConfigureRouter(r chi.Router) {
	redirectUrl, err := url.Parse(sc.oauthConfig.RedirectURL)
	if err != nil {
		panic(err.Error())
	}

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			noAuthPaths := map[string]struct{}{
				redirectUrl.Path:          {},
				sc.oauthConfig.LogoutPath: {},
				"/api/v1/liveness":        {},
				"/api/v1/readiness":       {},
			}

			if _, found := noAuthPaths[r.URL.Path]; !found {
				sc.oauthCodeGrantMiddleware(next).ServeHTTP(w, r)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	})

	r.Get(redirectUrl.Path, sc.codeExchange)
	r.Get(sc.oauthConfig.LogoutPath, sc.logout)
}

func (sc *SecurityConfiguration) logout(w http.ResponseWriter, r *http.Request) {
	session, err := sc.store.Get(r, sessionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Options.MaxAge = -1

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, sc.oauthConfig.HomePath+`?refresh=`+strconv.FormatInt(time.Now().Unix(), 10), http.StatusSeeOther)
}

func (sc *SecurityConfiguration) oauthCodeGrantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := sc.store.Get(r, sessionName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !isAuthenticated(session) {
			oauthStateString, err := generateRandomString(32)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			session.Values[oauthStateKey] = oauthStateString
			session.Values[redirectBackKey] = r.URL.String()
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			url := sc.oauthConfig.AuthCodeURL(oauthStateString, oauth2.ApprovalForce)
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
			return
		}

		next.ServeHTTP(w, r.WithContext(controller.SetContextParam(controller.UserIDKey, session.Values[controller.UserIDKey], r.Context())))
	})
}

func (sc *SecurityConfiguration) codeExchange(w http.ResponseWriter, r *http.Request) {
	session, err := sc.store.Get(r, sessionName)
	if err != nil {
		log.Println(err)
		return
	}
	state := session.Values[oauthStateKey].(string)

	receivedState := r.FormValue(oauthStateKey)
	if receivedState != state {
		log.Println("invalid oauth state ", err)
		return
	}

	code := r.FormValue("code")
	token, err := sc.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Println("oauth error:", err)
		return
	}

	log.Println("Access Token:", token.AccessToken)

	err = sc.saveUserInSession(r, token, session)
	if err != nil {
		log.Println("error retrieving user info:", err)
		return
	}

	now := time.Now().Unix()
	validityInSeconds := int64(token.Expiry.Sub(time.Unix(now, 0)).Seconds())
	session.Options.MaxAge = int(float64(validityInSeconds) * 0.95)

	session.Values[controller.UserIDKey] = controller.GetContextParam[int64](controller.UserIDKey, r.Context())

	err = sc.store.Save(r, w, session)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	redirectBack := session.Values[redirectBackKey].(string)
	if redirectBack == "" {
		redirectBack = sc.oauthConfig.HomePath
	}
	http.Redirect(w, r, redirectBack, http.StatusTemporaryRedirect)
}

func (sc *SecurityConfiguration) saveUserInSession(r *http.Request, token *oauth2.Token, session *sessions.Session) error {
	userInfo, err := sc.getUserInfoFromToken(r, token)
	if err != nil {
		return err
	}

	user, err := sc.userDAO.GetByEmail(userInfo.Email)
	if errors.Is(err, sql.ErrNoRows) {
		user = &model.User{}
		user.FirstName.Scan(userInfo.FirstName)
		user.LastName.Scan(userInfo.LastName)
		user.Name.Scan(userInfo.Name)
		user.PictureURL.Scan(userInfo.PictureURL)
		user.Email.Scan(userInfo.Email)

		user, err = sc.userDAO.Create(user)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	*r = *r.WithContext(controller.SetContextParam(controller.UserIDKey, user.ID.Int64, r.Context()))
	return nil
}

func (sc *SecurityConfiguration) getUserInfoFromToken(r *http.Request, token *oauth2.Token) (*UserInfo, error) {
	client := sc.oauthConfig.Client(context.Background(), token)
	resp, err := client.Get(sc.oauthConfig.UserEndpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo UserInfo
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func isAuthenticated(session *sessions.Session) bool {
	v, ok := session.Values[controller.UserIDKey].(int64)
	return ok && v != 0
}

func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}
