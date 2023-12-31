package controller

import (
	"net/http"
	"online-store/dao"
	"online-store/model"

	"github.com/go-chi/chi/v5"
)

func newUserRouter() chi.Router {
	userController := newUserController()
	r := chi.NewRouter()

	r.Get("/", ControllerHandler(userController.getAll))
	r.Get("/me", ControllerHandler(userController.getLoggedInUser))

	r.Route("/{id}", func(r chi.Router) {
		r.Use(numericPathVariableExtractor("id"))
		r.Get("/", ControllerHandler(userController.getByID))
		r.Put("/", ControllerHandler(userController.put))
	})

	return r
}

type userController struct {
	userDAO *dao.UserDAO
}

func newUserController() *userController {
	return &userController{
		userDAO: dao.NewUserDAO(),
	}
}

func (u *userController) getByID(r *http.Request) (*HTTPResponse[*model.User], error) {
	userId := GetContextParam[int64]("id", r.Context())

	user, err := u.userDAO.GetByID(userId)
	if err != nil {
		return nil, &HTTPError{Code: http.StatusNotFound, Message: "User not found", Err: err}
	}

	return NewOKResponse(user), nil
}

func (u *userController) getAll(r *http.Request) (*HTTPResponse[[]*model.User], error) {
	users, err := u.userDAO.GetAll()
	if err != nil {
		return nil, &HTTPError{Code: http.StatusInternalServerError, Message: "Cannot get users", Err: err}
	}

	return NewOKResponse(users), nil

}

func (u *userController) getLoggedInUser(r *http.Request) (*HTTPResponse[*model.User], error) {
	userId := GetContextParam[int64](UserIDKey, r.Context())

	users, err := u.userDAO.GetByID(userId)
	if err != nil {
		return nil, &HTTPError{Code: http.StatusInternalServerError, Message: "Cannot get users", Err: err}
	}

	return NewOKResponse(users), nil

}

func (u *userController) put(r *http.Request) (*HTTPResponse[*model.User], error) {
	userId := GetContextParam[int64](UserIDKey, r.Context())

	if userId != GetContextParam[int64]("id", r.Context()) {
		return nil, &HTTPError{Code: http.StatusBadRequest, Message: "Cannot update user", Err: nil}
	}

	user, err := jsonUnmarshalBody[model.User](r)
	if err != nil {
		return nil, err
	}
	user.ID.Scan(userId)

	user, err = u.userDAO.Update(user)
	if err != nil {
		return nil, &HTTPError{Code: http.StatusInternalServerError, Message: "Cannot update user", Err: nil}
	}

	return NewOKResponse(user), nil
}
