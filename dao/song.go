package dao

import (
	"mouban/common"
	"mouban/model"

	"github.com/sirupsen/logrus"
)

func UpsertSong(song *model.Song) {
	logrus.WithField("upsert", "song").Infoln("upsert song", song.DoubanId, song.Title)
	data := &model.Song{}
	common.Db.Where("douban_id = ? ", song.DoubanId).Assign(song).FirstOrCreate(data)
}

func UpdateSongThumbnail(doubanId uint64, thumbnail string) {
	common.Db.Model(&model.Song{}).Where("douban_id = ?", doubanId).Update("thumbnail", thumbnail)
}

func CreateSongNx(song *model.Song) bool {
	data := &model.Song{}
	inserted := common.Db.Where("douban_id = ? ", song.DoubanId).Attrs(song).FirstOrCreate(data).RowsAffected > 0
	if inserted {
		logrus.Infoln("create song", song.DoubanId, song.Title)
	}
	return inserted
}

func GetSongDetail(doubanId uint64) *model.Song {
	song := &model.Song{}
	common.Db.Where("douban_id = ? ", doubanId).Find(song)
	if song.ID == 0 {
		return nil
	}
	return song
}

func ListSongBrief(doubanIds *[]uint64) *[]model.Song {
	var songs *[]model.Song
	common.Db.Omit("intro", "track_list").Where("douban_id IN ? ", *doubanIds).Find(&songs)
	return songs
}
