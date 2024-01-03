package dao

import "online-store/model"

const (
	selectComments = `
		SELECT c.id, c.product_id, c.comment, c.created_at,
			u.id, u.name, u.first_name, u.last_name, u.picture_url, u.email, u.created_at,
			a.id, a.city, a.country, a.address, a.postal_code
		FROM comments c
		JOIN users u ON u.id = c.user_id
		LEFT JOIN addresses a ON a.id = u.id
	`

	selectCommentsByProductID = selectComments + " WHERE product_id = $1"

	insertComment = `
		INSERT INTO comments(user_id, product_id, comment)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	deleteComment = `
		DELETE FROM comments 
		WHERE id = $1
	`
)

type CommentDAO struct {
	dao     *DAO
	qe      queryExecutor
	userDAO *UserDAO
}

func NewCommentDAO() *CommentDAO {
	return newCommentDAO(GetDAO().db)
}

func newCommentDAO(qe queryExecutor) *CommentDAO {
	return &CommentDAO{
		dao:     GetDAO(),
		qe:      qe,
		userDAO: NewUserDAO(),
	}
}

func (c *CommentDAO) GetByProductID(productID int64) ([]*model.Comment, error) {
	return executeMultiRowQuery(c.qe, scanComment,
		selectCommentsByProductID, productID)
}

func (c *CommentDAO) Create(comment *model.Comment) (*model.Comment, error) {
	comment, err := executeSingleRowQuery(c.qe, propertyScanner(comment, &comment.ID, &comment.CreatedAt),
		insertComment, comment.User.ID, comment.ProductID, comment.Comment)
	if err != nil {
		return nil, err
	}

	user, err := c.userDAO.GetByID(comment.User.ID.Int64)
	if err != nil {
		return nil, err
	}

	comment.User = *user
	return comment, nil
}

func (c *CommentDAO) Delete(id int64) error {
	return executeNoRowsQuery(c.qe, deleteComment, id)
}

func scanComment(row rowScanner) (*model.Comment, error) {
	var comment model.Comment
	return propertyScanner(&comment,
		&comment.ID, &comment.ProductID, &comment.Comment, &comment.CreatedAt,
		&comment.User.ID, &comment.User.Name, &comment.User.FirstName, &comment.User.LastName, &comment.User.PictureURL, &comment.User.Email, &comment.User.CreatedAt,
		&comment.User.Address.ID, &comment.User.Address.City, &comment.User.Address.Country, &comment.User.Address.Address, &comment.User.Address.PostalCode)(row)
}
