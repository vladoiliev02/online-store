package controller

import (
	"net/http"
	"online-store/dao"
	"online-store/model"

	"github.com/go-chi/chi/v5"
)

func newOrderRouter() chi.Router {
	orderController := newOrderController()
	r := chi.NewRouter()

	r.Get("/", ControllerHandler(orderController.getAll))
	r.Post("/", ControllerHandler(orderController.post))
	r.Route("/{orderId}", func(r chi.Router) {
		r.Use(numericPathVariableExtractor("orderId"))
		r.Get("/", ControllerHandler(orderController.getByID))
		r.Put("/", ControllerHandler(orderController.put))
	})
	r.Get("/cart", ControllerHandler(orderController.getCart))
	r.Mount("/cart/items", newItemRouter(orderController))

	return r
}

type orderController struct {
	orderDao *dao.OrderDAO
}

func newOrderController() *orderController {
	return &orderController{
		orderDao: dao.NewOrderDAO(),
	}
}

func (o *orderController) getByID(r *http.Request) (*HTTPResponse[*model.Order], error) {
	id := GetContextParam[int64]("id", r.Context())

	order, err := o.orderDao.GetByID(id)
	if err != nil {
		return nil, &HTTPError{
			Code:    http.StatusNotFound,
			Message: "Order not found",
			Err:     err,
		}
	}

	return NewOKResponse(order), nil
}

func (o *orderController) getCart(r *http.Request) (*HTTPResponse[*model.Order], error) {
	userID := GetContextParam[int64](UserIDKey, r.Context())

	cart, err := o.orderDao.GetCart(userID)
	if err != nil {
		return nil, &HTTPError{
			Code:    http.StatusNotFound,
			Message: "Cart not found",
			Err:     err,
		}
	}

	return NewOKResponse(cart), nil
}

func (o *orderController) getAll(r *http.Request) (*HTTPResponse[[]*model.Order], error) {
	userID := GetContextParam[int64]("userID", r.Context())

	statusStr := getQueryParam(r, "status")
	var err error
	var status int64
	if statusStr != "" {
		status, err = toInt(statusStr)
		if err != nil || !model.IsValidOrderStatus(model.OrderStatus(status)) {
			return nil, &HTTPError{Code: http.StatusBadRequest, Message: "Invalid order status", Err: err}
		}
	}

	var order []*model.Order
	if status == 0 {
		order, err = o.orderDao.GetByUserID(userID)
	} else {
		order, err = o.orderDao.GetByUserIDAndStatus(userID, model.OrderStatus(status))
	}

	if err != nil {
		return nil, &HTTPError{Code: http.StatusNotFound, Message: "Order not found", Err: err}
	}
	return NewOKResponse(order), nil
}

func (o *orderController) post(r *http.Request) (*HTTPResponse[*model.Order], error) {
	order, err := jsonUnmarshalBody[model.Order](r)
	if err != nil {
		return nil, err
	}

	if err := model.ValidateOrder(order, false); err != nil {
		return nil, &HTTPError{Code: http.StatusBadRequest, Message: "Invalid order", Err: err}
	}

	order, err = o.orderDao.Create(order)
	if err != nil {
		return nil, &HTTPError{Code: http.StatusInternalServerError, Message: "Order creation error", Err: err}
	}

	return NewResponse(http.StatusCreated, order), nil
}

func (o *orderController) put(r *http.Request) (*HTTPResponse[*model.Order], error) {
	id := GetContextParam[int64]("id", r.Context())

	newOrder, err := jsonUnmarshalBody[model.Order](r)
	if err != nil {
		return nil, &HTTPError{Code: http.StatusBadRequest, Message: "Invalid order", Err: err}
	}

	newOrder.ID.Scan(id)
	if err := model.ValidateOrder(newOrder, true); err != nil {
		return nil, &HTTPError{Code: http.StatusBadRequest, Message: "Invalid order", Err: err}
	}

	result, err := o.orderDao.Update(newOrder)
	if err != nil {
		return nil, &HTTPError{Code: http.StatusInternalServerError, Message: "Order update error", Err: err}
	}

	return NewOKResponse(result), nil
}

func newItemRouter(orderController *orderController) chi.Router {
	itemController := newItemController(orderController)
	r := chi.NewRouter()

	r.Get("/", ControllerHandler(itemController.getAll))
	r.Post("/", ControllerHandler(itemController.post))

	r.Route("/{itemID}", func(r chi.Router) {
		r.Use(numericPathVariableExtractor("itemID"))
		r.Delete("/", ControllerHandler(itemController.delete))
	})

	return r
}

type itemController struct {
	orderDAO        *dao.OrderDAO
	orderController *orderController
}

func newItemController(orderController *orderController) *itemController {
	return &itemController{
		orderDAO:        dao.NewOrderDAO(),
		orderController: orderController,
	}
}

func (i *itemController) getAll(r *http.Request) (*HTTPResponse[*model.Order], error) {
	cartResponse, err := i.orderController.getCart(r)
	if err != nil {
		return nil, err
	}

	cart, err := i.orderDAO.LoadItems(cartResponse.Body)
	if err != nil {
		return nil, &HTTPError{Code: http.StatusInternalServerError, Message: "Cannot load items", Err: err}
	}

	return NewOKResponse(cart), nil
}

func (i *itemController) post(r *http.Request) (*HTTPResponse[*model.Item], error) {
	userID := GetContextParam[int64](UserIDKey, r.Context())
	item, err := jsonUnmarshalBody[model.Item](r)
	if err != nil {
		return nil, err
	}

	if err := model.ValidateItem(item); err != nil {
		return nil, &HTTPError{Code: http.StatusBadRequest, Message: "Invalid item", Err: err}
	}

	item, err = i.orderDAO.AddItem(userID, item)
	if err != nil {
		return nil, &HTTPError{Code: http.StatusInternalServerError, Message: "Cannot add item", Err: err}
	}

	return NewOKResponse(item), nil
}

func (i *itemController) delete(r *http.Request) (*HTTPResponse[any], error) {
	userID := GetContextParam[int64](UserIDKey, r.Context())
	itemID := GetContextParam[int64]("itemID", r.Context())

	err := i.orderDAO.RemoveItem(userID, itemID)
	if err != nil {
		return nil, &HTTPError{Code: http.StatusInternalServerError, Message: "Cannot remove item from cart", Err: err}
	}

	return NewOKResponse[any](nil), nil
}
