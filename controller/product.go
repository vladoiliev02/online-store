package controller

import (
	"log"
	"net/http"
	"online-store/dao"
	"online-store/model"

	"github.com/go-chi/chi/v5"
)

const (
	productIdCtxKey = "productId"
	commentIdCtxKey = "commentId"
	imageIdCtxKey   = "imageId"
)

func newProductRouter() chi.Router {
	productController := newProductController()
	r := chi.NewRouter()

	r.Get("/", ControllerHandler(productController.getAll))
	r.Post("/", ControllerHandler(productController.post))

	r.Route("/{productId}", func(r chi.Router) {
		r.Use(numericPathVariableExtractor(productIdCtxKey))
		r.Get("/", ControllerHandler(productController.getById))
		r.Put("/", ControllerHandler(productController.put))
		r.Patch("/", ControllerHandler(productController.rateProduct))
		r.Mount("/comments", newCommentRouter())
		r.Mount("/images", newImageRouter())
	})

	return r
}

type productController struct {
	productDAO *dao.ProductDAO
}

func newProductController() *productController {
	return &productController{
		productDAO: dao.NewProductDAO(),
	}
}

func (p *productController) getById(r *http.Request) (*HTTPResponse[*model.Product], error) {
	id := GetContextParam[int64](productIdCtxKey, r.Context())

	product, err := p.productDAO.GetByID(id)
	if err != nil {
		return nil, &HTTPError{Code: http.StatusNotFound, Message: "Product not found", Err: err}
	}

	return NewOKResponse(product), nil
}

func (p *productController) getAll(r *http.Request) (*HTTPResponse[[]*model.Product], error) {
	page, pageSize, err := getPageAndPageSize(r)
	if err != nil {
		return nil, err
	}

	name := getQueryParam(r, "name")
	var result []*model.Product
	if name != "" {
		log.Println("searching by name:", name)
		result, err = p.productDAO.GetByNameLike(name, page, pageSize)
	} else {
		result, err = p.productDAO.GetAll(page, pageSize)
	}

	if err != nil {
		return nil, &HTTPError{Code: http.StatusNotFound, Message: "Products not found", Err: err}
	}

	return NewOKResponse(result), nil
}

func (p *productController) post(r *http.Request) (*HTTPResponse[*model.Product], error) {
	product, err := jsonUnmarshalBody[model.Product](r)
	if err != nil {
		return nil, err
	}

	err = model.ValidateProduct(product, false)
	if err != nil {
		return nil, &HTTPError{
			Code:    http.StatusBadRequest,
			Message: "Invalid product",
			Err:     err,
		}
	}

	product, err = p.productDAO.Create(product)
	if err != nil {
		return nil, &HTTPError{
			Code:    http.StatusInternalServerError,
			Message: "Product creation error",
			Err:     err,
		}
	}

	return NewResponse(http.StatusOK, product), nil
}

func (p *productController) put(r *http.Request) (*HTTPResponse[*model.Product], error) {
	id := GetContextParam[int64](productIdCtxKey, r.Context())

	product, err := jsonUnmarshalBody[model.Product](r)
	if err != nil {
		return nil, err
	}

	product.ID.Scan(id)
	err = model.ValidateProduct(product, true)
	if err != nil {
		return nil, &HTTPError{Code: http.StatusBadRequest, Message: "Invalid product", Err: err}
	}

	product, err = p.productDAO.Update(product)
	if err != nil {
		return nil, &HTTPError{Code: http.StatusBadRequest, Message: "Could not update product", Err: err}
	}

	return NewOKResponse(product), nil
}

func (p *productController) rateProduct(r *http.Request) (*HTTPResponse[*model.Product], error) {
	id := GetContextParam[int64](productIdCtxKey, r.Context())

	rating, err := jsonUnmarshalBody[model.Rating](r)
	if err != nil {
		return nil, err
	}

	rating.UserID.Scan(GetContextParam[int64](UserIDKey, r.Context()))
	rating.ProductID.Scan(id)

	if err := model.ValidateRating(rating); err != nil {
		return nil, &HTTPError{Code: http.StatusBadRequest, Message: "Invalid rating", Err: err}
	}

	product, err := p.productDAO.AddRating(rating)
	if err != nil {
		return nil, &HTTPError{Code: http.StatusInternalServerError, Message: "Cannot add rating", Err: err}
	}

	return NewOKResponse(product), nil
}

func newCommentRouter() chi.Router {
	commentController := newCommentController()
	r := chi.NewRouter()

	r.Get("/", ControllerHandler(commentController.getAll))
	r.Post("/", ControllerHandler(commentController.post))

	r.Route("/{commentId}", func(r chi.Router) {
		r.Use(numericPathVariableExtractor(commentIdCtxKey))
		r.Delete("/", ControllerHandler(commentController.delete))
	})

	return r
}

type commentController struct {
	commentDAO *dao.CommentDAO
}

func newCommentController() *commentController {
	return &commentController{
		commentDAO: dao.NewCommentDAO(),
	}
}

func (c *commentController) getAll(r *http.Request) (*HTTPResponse[[]*model.Comment], error) {
	productId := GetContextParam[int64](productIdCtxKey, r.Context())

	comments, err := c.commentDAO.GetByProductID(productId)
	if err != nil {
		return nil, &HTTPError{Code: http.StatusNotFound, Message: "Comments not found", Err: err}
	}

	return NewOKResponse(comments), nil
}

func (c *commentController) post(r *http.Request) (*HTTPResponse[*model.Comment], error) {
	productId := GetContextParam[int64](productIdCtxKey, r.Context())
	comment, err := jsonUnmarshalBody[model.Comment](r)
	if err != nil {
		return nil, err
	}

	comment.User.ID.Scan(GetContextParam[int64](UserIDKey, r.Context()))
	comment.ProductID.Scan(productId)

	if err := model.ValidateComment(comment, false); err != nil {
		return nil, &HTTPError{Code: http.StatusBadRequest, Message: "Invalid comment", Err: err}
	}

	comment, err = c.commentDAO.Create(comment)
	if err != nil {
		return nil, &HTTPError{Code: http.StatusInternalServerError, Message: "Cannot create comment", Err: err}
	}

	return NewResponse(http.StatusCreated, comment), nil
}

func (c *commentController) delete(r *http.Request) (*HTTPResponse[any], error) {
	id := GetContextParam[int64](commentIdCtxKey, r.Context())

	err := c.commentDAO.Delete(id)
	if err != nil {
		return nil, &HTTPError{Code: http.StatusInternalServerError, Message: "Cannot delete comment", Err: err}
	}

	return NewStatusResponse[any](http.StatusOK), nil
}

func getPageAndPageSize(r *http.Request) (int, int, error) {
	page, err := getNumericQueryParam(r, "page")
	if err != nil {
		return 0, 0, &HTTPError{Code: http.StatusBadRequest, Message: "Invalid value for page", Err: err}
	}

	pageSize, err := getNumericQueryParam(r, "pageSize")
	if err != nil {
		return 0, 0, &HTTPError{Code: http.StatusBadRequest, Message: "Invalid value for pageSize", Err: err}
	}

	return int(page), int(pageSize), nil
}

type imageController struct {
	imageDAO *dao.ImageDAO
}

func newImageController() *imageController {
	return &imageController{
		imageDAO: dao.NewImageDAO(),
	}
}

func newImageRouter() chi.Router {
	imageController := newImageController()
	r := chi.NewRouter()

	r.Get("/", ControllerHandler(imageController.getByProductId))
	r.Post("/", ControllerHandler(imageController.post))

	r.Route("/{imageId}", func(r chi.Router) {
		r.Use(numericPathVariableExtractor(imageIdCtxKey))
		r.Delete("/", ControllerHandler(imageController.delete))
	})

	return r
}

func (i *imageController) getByProductId(r *http.Request) (*HTTPResponse[[]*model.Image], error) {
	productId := GetContextParam[int64](productIdCtxKey, r.Context())

	images, err := i.imageDAO.GetByProductID(productId)
	if err != nil {
		return nil, &HTTPError{Code: http.StatusNotFound, Message: "Cannot find images", Err: err}
	}

	return NewOKResponse(images), nil
}

func (i *imageController) post(r *http.Request) (*HTTPResponse[*model.Image], error) {
	productId := GetContextParam[int64](productIdCtxKey, r.Context())

	image, err := jsonUnmarshalBody[model.Image](r)
	if err != nil {
		return nil, err
	}
	image.ProductID.Scan(productId)

	if err := model.ValidateImage(image); err != nil {
		return nil, &HTTPError{Code: http.StatusBadRequest, Message: "Invalid image", Err: err}
	}

	image, err = i.imageDAO.Create(image)
	if err != nil {
		return nil, &HTTPError{Code: http.StatusInternalServerError, Message: "Cannot save image", Err: err}
	}

	return NewOKResponse(image), nil
}

func (i *imageController) delete(r *http.Request) (*HTTPResponse[any], error) {
	id := GetContextParam[int64](imageIdCtxKey, r.Context())

	err := i.imageDAO.Delete(id)
	if err != nil {
		return nil, &HTTPError{Code: http.StatusNotFound, Message: "Cannot delete image", Err: err}
	}

	return NewStatusResponse[any](http.StatusOK), nil
}
