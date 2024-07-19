package dao

import (
	"mouban/common"
	"mouban/model"

	"github.com/sirupsen/logrus"
)

func UpsertBook(book *model.Book) {
	logrus.WithField("upsert", "book").Infoln("upsert book", book.DoubanId, book.Title)
	data := &model.Book{}
	common.Db.Where("douban_id = ? ", book.DoubanId).Assign(book).FirstOrCreate(data)
}

func UpdateBookThumbnail(doubanId uint64, thumbnail string) {
	common.Db.Model(&model.Book{}).Where("douban_id = ?", doubanId).Update("thumbnail", thumbnail)
}

func CreateBookNx(book *model.Book) bool {
	data := &model.Book{}
	inserted := common.Db.Where("douban_id = ? ", book.DoubanId).Attrs(book).FirstOrCreate(data).RowsAffected > 0
	if inserted {
		logrus.Infoln("create book", book.DoubanId, book.Title)
	}
	return inserted
}

func GetBookDetail(doubanId uint64) *model.Book {
	book := &model.Book{}
	common.Db.Where("douban_id = ? ", doubanId).Find(book)
	if book.ID == 0 {
		return nil
	}
	return book
}

func ListBookBrief(doubanIds *[]uint64) *[]model.Book {
	var books *[]model.Book
	common.Db.Omit("serial", "isbn", "framing", "page", "intro").Where("douban_id IN ? ", *doubanIds).Find(&books)
	return books
}
