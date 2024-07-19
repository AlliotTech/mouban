package dao

import (
	"mouban/common"
	"mouban/model"

	"github.com/sirupsen/logrus"
)

func UpsertMovie(movie *model.Movie) {
	logrus.WithField("upsert", "movie").Infoln("upsert movie", movie.DoubanId, movie.Title)
	data := &model.Movie{}
	common.Db.Where("douban_id = ? ", movie.DoubanId).Assign(movie).FirstOrCreate(data)
}

func UpdateMovieThumbnail(doubanId uint64, thumbnail string) {
	common.Db.Model(&model.Movie{}).Where("douban_id = ?", doubanId).Update("thumbnail", thumbnail)
}

func CreateMovieNx(movie *model.Movie) bool {
	data := &model.Movie{}
	inserted := common.Db.Where("douban_id = ? ", movie.DoubanId).Attrs(movie).FirstOrCreate(data).RowsAffected > 0
	if inserted {
		logrus.Infoln("create movie", movie.DoubanId, movie.Title)
	}
	return inserted
}

func GetMovieDetail(doubanId uint64) *model.Movie {
	movie := &model.Movie{}
	common.Db.Where("douban_id = ? ", doubanId).Find(movie)
	if movie.ID == 0 {
		return nil
	}
	return movie
}

func ListMovieBrief(doubanIds *[]uint64) *[]model.Movie {
	var movies *[]model.Movie
	common.Db.Omit("intro").Where("douban_id IN ? ", *doubanIds).Find(&movies)
	return movies
}
