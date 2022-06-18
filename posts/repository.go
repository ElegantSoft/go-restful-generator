package posts

//
//import (
//	"github.com/ElegantSoft/shabahy/common"
//	"github.com/ElegantSoft/shabahy/db"
//)
//
//type Repository struct {
//	crud *common.CrudRepository
//}
//
//func (r *Repository) Find(id uint) (error, interface{}) {
//	var itemToFind Interest
//	return r.crud.Find(id, &itemToFind)
//}
//
//func (r *Repository) Create(item *Interest) (error, interface{}) {
//	return r.crud.Create(item)
//}
//
//func (r *Repository) Update(item *Interest, id uint) error {
//	return r.crud.Update(id, &item)
//}
//
//func (r *Repository) Delete(id uint) error {
//	return r.crud.Delete(id)
//}
//
//func (r *Repository) FindOne(interest *Interest, dest *Interest) error {
//	return db.DB.Where(interest).First(dest).Error
//}
//
//func NewRepository(crud *common.CrudRepository) *Repository {
//	return &Repository{
//		crud: crud,
//	}
//}
//
//func InitRepository() *Repository {
//	return &Repository{
//		crud: common.NewCrudRepository("categories"),
//	}
//}
