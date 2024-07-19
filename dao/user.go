package dao

import (
	"mouban/common"
	"mouban/model"
	"time"

	"github.com/sirupsen/logrus"
)

func UpsertUser(user *model.User) {
	logrus.WithField("upsert", "user").Infoln("upsert user", user.DoubanUid, user.Name)
	data := &model.User{}
	common.Db.Where("douban_uid = ? ", user.DoubanUid).Assign(user).FirstOrCreate(data)
}

func RefreshUser(user *model.User) {
	logrus.Infoln("refresh user", user.DoubanUid, user.Name)
	common.Db.Model(&model.User{}).
		Where("douban_uid = ? ", user.DoubanUid).
		Updates(model.User{CheckAt: time.Unix(0, 0), SyncAt: time.Unix(0, 0), PublishAt: time.Unix(0, 0)})
}

func GetUser(doubanUid uint64) *model.User {
	if doubanUid == 0 {
		return nil
	}
	user := &model.User{}
	common.Db.Where("douban_uid = ? ", doubanUid).Find(user)
	if user.ID == 0 {
		return nil
	}
	return user
}

func GetUserByDomain(domain string) *model.User {
	if domain == "" {
		return nil
	}
	user := &model.User{}
	common.Db.Where("domain = ? ", domain).Find(user)
	if user.ID == 0 {
		return nil
	}
	return user
}
