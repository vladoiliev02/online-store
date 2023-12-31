package dao

import "online-store/model"

const (
	selectImages = `
		SELECT id, product_id, data
		FROM product_images
	`

	selectByProductId = selectImages +
		"WHERE product_id = $1"

	insertImage = `
		INSERT INTO images (product_id, data)
		VALUES ($1, $2)
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

func (i *ImageDAO) GetByProductID(productID int64) ([]*model.Image, error) {
	return executeMultiRowQuery(i.qe,
		scanImage,
		selectByProductId, productID)
}

func (i *ImageDAO) Create(image *model.Image) (*model.Image, error) {
	return executeSingleRowQuery(i.qe,
		propertyScanner(image, &image.ID),
		insertImage, image.ProductID, image.Data)
}

func (i *ImageDAO) Delete(id int64) error {
	return executeNoRowsQuery(i.qe,
		deleteImage, id)
}

func scanImage(row rowScanner) (*model.Image, error) {
	var image model.Image
	return propertyScanner(&image, &image.ID, &image.ProductID, &image.Data)(row)
}
