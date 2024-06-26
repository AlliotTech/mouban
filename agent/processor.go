package agent

import (
	"mouban/consts"
	"mouban/crawl"
	"mouban/dao"
	"mouban/model"
	"mouban/util"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func processItem(t uint8, doubanId uint64) {
	switch t {
	case consts.TypeBook.Code:
		processBook(doubanId)
	case consts.TypeMovie.Code:
		processMovie(doubanId)
	case consts.TypeGame.Code:
		processGame(doubanId)
	case consts.TypeSong.Code:
		processSong(doubanId)
	}
}

func processBook(doubanId uint64) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorln("process book panic", doubanId, r, "=>", util.GetCurrentGoroutineStack())
		}
	}()
	book, rating, newUser, newItems, err := crawl.Book(doubanId)

	processDiscoverUser(newUser)
	processDiscoverItem(newItems, consts.TypeBook)

	if err != nil {
		dao.ChangeScheduleResult(doubanId, consts.TypeBook.Code, consts.ScheduleInvalid.Code)
		panic(err)
	}

	book.Thumbnail = crawl.Storage(book.Thumbnail)
	dao.UpsertBook(book)
	dao.UpsertRating(rating)

	dao.ChangeScheduleResult(doubanId, consts.TypeBook.Code, consts.ScheduleReady.Code)
}

func processMovie(doubanId uint64) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorln("process movie panic", doubanId, r, "=>", util.GetCurrentGoroutineStack())
		}
	}()
	movie, rating, newUser, newItems, err := crawl.Movie(doubanId)

	processDiscoverUser(newUser)
	processDiscoverItem(newItems, consts.TypeMovie)

	if err != nil {
		dao.ChangeScheduleResult(doubanId, consts.TypeMovie.Code, consts.ScheduleInvalid.Code)
		panic(err)
	}
	movie.Thumbnail = crawl.Storage(movie.Thumbnail)
	dao.UpsertMovie(movie)
	dao.UpsertRating(rating)

	dao.ChangeScheduleResult(doubanId, consts.TypeMovie.Code, consts.ScheduleReady.Code)
}

func processGame(doubanId uint64) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorln("process game panic", doubanId, r, "=>", util.GetCurrentGoroutineStack())
		}
	}()

	game, rating, newUser, newItems, err := crawl.Game(doubanId)

	processDiscoverUser(newUser)
	processDiscoverItem(newItems, consts.TypeGame)

	if err != nil {
		dao.ChangeScheduleResult(doubanId, consts.TypeGame.Code, consts.ScheduleInvalid.Code)
		panic(err)
	}
	game.Thumbnail = crawl.Storage(game.Thumbnail)
	dao.UpsertGame(game)
	dao.UpsertRating(rating)

	dao.ChangeScheduleResult(doubanId, consts.TypeGame.Code, consts.ScheduleReady.Code)
}

func processSong(doubanId uint64) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorln("process song panic", doubanId, r, "=>", util.GetCurrentGoroutineStack())
		}
	}()

	song, rating, newUser, newItems, err := crawl.Song(doubanId)

	processDiscoverUser(newUser)
	processDiscoverItem(newItems, consts.TypeSong)

	if err != nil {
		dao.ChangeScheduleResult(doubanId, consts.TypeSong.Code, consts.ScheduleInvalid.Code)
		panic(err)
	}
	song.Thumbnail = crawl.Storage(song.Thumbnail)
	dao.UpsertSong(song)
	dao.UpsertRating(rating)

	dao.ChangeScheduleResult(doubanId, consts.TypeSong.Code, consts.ScheduleReady.Code)
}

func processDiscoverUser(newUsers *[]string) {
	if newUsers == nil {
		return
	}
	level := viper.GetInt("agent.discover.level")
	if level == 0 {
		return
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorln("process discover user panic", r, "=>", util.GetCurrentGoroutineStack())
			}
		}()
		totalFound := len(*newUsers)
		newFound := 0
		for _, idOrDomain := range *newUsers {
			id, err := strconv.ParseUint(idOrDomain, 10, 64)
			if err != nil {
				if level == 1 {
					continue
				}
				user := dao.GetUserByDomain(idOrDomain)
				if user == nil {
					id = crawl.UserId(idOrDomain)
				}
			}
			if id > 0 {
				added := dao.CreateScheduleNx(id, consts.TypeUser.Code, consts.ScheduleCanCrawl.Code, consts.ScheduleUnready.Code)
				if added {
					newFound += 1
				}
			}
		}
		if newFound > 0 {
			logrus.Infoln("(", newFound, "/", totalFound, ") users discovered")
		}
	}()
}

func processDiscoverItem(newItems *[]uint64, t consts.Type) {
	if newItems == nil || len(*newItems) == 0 {
		return
	}
	level := viper.GetInt("agent.discover.level")
	if level == 0 {
		return
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorln("process discover item panic", r, "=>", util.GetCurrentGoroutineStack())
			}
		}()
		totalFound := len(*newItems)
		newFound := 0
		for _, doubanId := range *newItems {
			added := dao.CreateScheduleNx(doubanId, t.Code, consts.ScheduleCanCrawl.Code, consts.ScheduleUnready.Code)
			if added {
				newFound += 1
			}
		}
		if newFound > 0 {
			logrus.Infoln("(", newFound, "/", totalFound, ")", t.Name, "discovered")
		}
	}()
}

func processUser(doubanUid uint64) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorln("process user panic", doubanUid, r, "=>", util.GetCurrentGoroutineStack())
		}
	}()

	userPublish, _ := crawl.UserPublish(doubanUid)
	rawUser := dao.GetUser(doubanUid)
	if rawUser != nil && rawUser.PublishAt.Equal(userPublish) {
		logrus.Infoln("user", doubanUid, "changed ->", false)
		rawUser.CheckAt = time.Now()
		dao.UpsertUser(rawUser)
		return
	}
	logrus.Infoln("user", doubanUid, "changed ->", true)

	//user
	user, err := crawl.UserOverview(doubanUid)
	if err != nil {
		dao.ChangeScheduleResult(doubanUid, consts.TypeUser.Code, consts.ScheduleInvalid.Code)
		panic(err)
	}

	// choose update type
	forceSyncAfter := time.Unix(0, 0)
	if rawUser != nil && rawUser.SyncAt.AddDate(1, 0, 0).After(time.Now()) {
		forceSyncAfter = rawUser.SyncAt
	}

	logrus.Infoln("user", doubanUid, "sync_after ->", forceSyncAfter)

	//book
	if user.BookDo+user.BookWish+user.BookCollect > 0 {
		syncCommentBook(user, forceSyncAfter)
	}

	//movie
	if user.MovieDo+user.MovieWish+user.MovieCollect > 0 {
		syncCommentMovie(user, forceSyncAfter)
	}

	//game
	if user.GameDo+user.GameWish+user.GameCollect > 0 {
		syncCommentGame(user, forceSyncAfter)
	}

	//song
	if user.SongDo+user.SongWish+user.SongCollect > 0 {
		syncCommentSong(user, forceSyncAfter)
	}

	user.CheckAt = time.Now()
	user.SyncAt = time.Now()

	user.Thumbnail = crawl.Storage(user.Thumbnail)
	dao.UpsertUser(user)
	dao.ChangeScheduleResult(doubanUid, consts.TypeUser.Code, consts.ScheduleReady.Code)
}

func syncCommentGame(user *model.User, forceSyncAfter time.Time) {
	comment, game, err := crawl.CommentGame(user, forceSyncAfter)
	if err != nil {
		panic(err)
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorln("sync comment game panic", r, "=>", util.GetCurrentGoroutineStack())
			}
		}()
		if forceSyncAfter.Unix() == 0 {
			newCommentIds := make(map[uint64]bool)
			for i := range *game {
				newCommentIds[(*game)[i].DoubanId] = true
			}
			oldCommentIds := dao.GetCommentIds(user.DoubanUid, consts.TypeGame.Code)
			for i := range *oldCommentIds {
				id := (*oldCommentIds)[i]
				if !newCommentIds[id] {
					dao.HideComment(user.DoubanUid, consts.TypeGame.Code, id)
				}
			}
		}

		for i := range *game {
			dao.UpsertComment(&(*comment)[i])

			item := &(*game)[i]
			newThumbnail := crawl.Storage(item.Thumbnail)
			thumbChange := item.Thumbnail != newThumbnail
			item.Thumbnail = newThumbnail
			added := dao.CreateGameNx(item)
			if added {
				dao.CreateScheduleNx(item.DoubanId, consts.TypeGame.Code, consts.ScheduleToCrawl.Code, consts.ScheduleUnready.Code)
			} else {
				if thumbChange {
					dao.UpdateGameThumbnail(item.DoubanId, item.Thumbnail)
				}
			}
		}
	}()
}

func syncCommentBook(user *model.User, forceSyncAfter time.Time) {
	comment, book, err := crawl.CommentBook(user, forceSyncAfter)
	if err != nil {
		panic(err)
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorln("sync comment book panic", r, "=>", util.GetCurrentGoroutineStack())
			}
		}()
		if forceSyncAfter.Unix() == 0 {
			newCommentIds := make(map[uint64]bool)
			for i := range *book {
				newCommentIds[(*book)[i].DoubanId] = true
			}
			oldCommentIds := dao.GetCommentIds(user.DoubanUid, consts.TypeBook.Code)
			for i := range *oldCommentIds {
				id := (*oldCommentIds)[i]
				if !newCommentIds[id] {
					dao.HideComment(user.DoubanUid, consts.TypeBook.Code, id)
				}
			}
		}
		for i := range *book {
			dao.UpsertComment(&(*comment)[i])

			item := &(*book)[i]
			newThumbnail := crawl.Storage(item.Thumbnail)
			thumbChange := item.Thumbnail != newThumbnail
			item.Thumbnail = newThumbnail
			added := dao.CreateBookNx(item)
			if added {
				dao.CreateScheduleNx(item.DoubanId, consts.TypeBook.Code, consts.ScheduleToCrawl.Code, consts.ScheduleUnready.Code)
			} else {
				if thumbChange {
					dao.UpdateBookThumbnail(item.DoubanId, item.Thumbnail)
				}
			}
		}
	}()
}

func syncCommentMovie(user *model.User, forceSyncAfter time.Time) {
	comment, movie, err := crawl.CommentMovie(user, forceSyncAfter)
	if err != nil {
		panic(err)
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorln("sync comment movie panic", r, "=>", util.GetCurrentGoroutineStack())
			}
		}()
		if forceSyncAfter.Unix() == 0 {
			newCommentIds := make(map[uint64]bool)
			for i := range *movie {
				newCommentIds[(*movie)[i].DoubanId] = true
			}
			oldCommentIds := dao.GetCommentIds(user.DoubanUid, consts.TypeMovie.Code)
			for i := range *oldCommentIds {
				id := (*oldCommentIds)[i]
				if !newCommentIds[id] {
					dao.HideComment(user.DoubanUid, consts.TypeMovie.Code, id)
				}
			}
		}
		for i := range *movie {
			dao.UpsertComment(&(*comment)[i])

			item := &(*movie)[i]
			newThumbnail := crawl.Storage(item.Thumbnail)
			thumbChange := item.Thumbnail != newThumbnail
			item.Thumbnail = newThumbnail
			added := dao.CreateMovieNx(item)
			if added {
				dao.CreateScheduleNx(item.DoubanId, consts.TypeMovie.Code, consts.ScheduleToCrawl.Code, consts.ScheduleUnready.Code)
			} else {
				if thumbChange {
					dao.UpdateMovieThumbnail(item.DoubanId, item.Thumbnail)
				}
			}
		}
	}()
}

func syncCommentSong(user *model.User, forceSyncAfter time.Time) {
	comment, song, err := crawl.CommentSong(user, forceSyncAfter)
	if err != nil {
		panic(err)
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorln("sync comment song panic", r, "=>", util.GetCurrentGoroutineStack())
			}
		}()
		if forceSyncAfter.Unix() == 0 {
			newCommentIds := make(map[uint64]bool)
			for i := range *song {
				newCommentIds[(*song)[i].DoubanId] = true
			}
			oldCommentIds := dao.GetCommentIds(user.DoubanUid, consts.TypeSong.Code)
			for i := range *oldCommentIds {
				id := (*oldCommentIds)[i]
				if !newCommentIds[id] {
					dao.HideComment(user.DoubanUid, consts.TypeSong.Code, id)
				}
			}
		}
		for i := range *song {
			item := &(*song)[i]
			dao.UpsertComment(&(*comment)[i])

			newThumbnail := crawl.Storage(item.Thumbnail)
			thumbChange := item.Thumbnail != newThumbnail
			item.Thumbnail = newThumbnail
			added := dao.CreateSongNx(item)
			if added {
				dao.CreateScheduleNx(item.DoubanId, consts.TypeSong.Code, consts.ScheduleToCrawl.Code, consts.ScheduleUnready.Code)
			} else {
				if thumbChange {
					dao.UpdateSongThumbnail(item.DoubanId, item.Thumbnail)
				}
			}
		}
	}()
}
