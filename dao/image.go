package dao

import (
	"github.com/vladoiliev02/online-store/model"
)

const (
	selectImages = `
		SELECT id, product_id, data, format
		FROM product_images
	`

	selectByProductId = selectImages + `
		WHERE product_id = $1
		LIMIT $2
	`

	insertImage = `
		INSERT INTO product_images (product_id, data, format)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	deleteImage = `
		DELETE
		FROM product_images
		WHERE id = $1
	`
)

type ImageDAO struct {
	dao *DAO
	qe  queryExecutor
}

func NewImageDAO() *ImageDAO {
	return newImageDAO(GetDAO().db)
}

func newImageDAO(qe queryExecutor) *ImageDAO {
	return &ImageDAO{
		dao: GetDAO(),
		qe:  qe,
	}
}

func (i *ImageDAO) GetByProductID(productID, limit int64) ([]*model.Image, error) {
	return executeMultiRowQuery(i.qe,
		scanImage,
		selectByProductId, productID, limit)
}

func (i *ImageDAO) Create(image *model.Image) (*model.Image, error) {
	return executeSingleRowQuery(i.qe,
		propertyScanner(image, &image.ID),
		insertImage, image.ProductID, image.Data, image.Format)
}

func (i *ImageDAO) Delete(id int64) error {
	return executeNoRowsQuery(i.qe,
		deleteImage, id)
}

func scanImage(row rowScanner) (*model.Image, error) {
	var image model.Image
	return propertyScanner(&image, &image.ID, &image.ProductID, &image.Data, &image.Format)(row)
}
