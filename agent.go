package main

import (
	"log"
	"mouban/consts"
	"mouban/crawl"
	"mouban/dao"
	"mouban/model"
	"mouban/util"
	"sync"
	"time"
)

func processBook(doubanId uint64) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r, " => ", util.GetCurrentGoroutineStack())
		}
	}()
	book, rating, err := crawl.Book(doubanId)

	if err != nil {
		dao.ChangeScheduleResult(doubanId, consts.TypeBook, consts.ScheduleResultInvalid)
		panic(err)
	}
	dao.UpsertBook(book)
	dao.UpsertRating(rating)

	dao.ChangeScheduleResult(doubanId, consts.TypeBook, consts.ScheduleResultReady)
}

func processMovie(doubanId uint64) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r, " => ", util.GetCurrentGoroutineStack())
		}
	}()
	movie, rating, err := crawl.Movie(doubanId)

	if err != nil {
		dao.ChangeScheduleResult(doubanId, consts.TypeMovie, consts.ScheduleResultInvalid)
		panic(err)
	}
	dao.UpsertMovie(movie)
	dao.UpsertRating(rating)

	dao.ChangeScheduleResult(doubanId, consts.TypeMovie, consts.ScheduleResultReady)
}

func processGame(doubanId uint64) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r, " => ", util.GetCurrentGoroutineStack())
		}
	}()

	game, rating, err := crawl.Game(doubanId)

	if err != nil {
		dao.ChangeScheduleResult(doubanId, consts.TypeGame, consts.ScheduleResultInvalid)
		panic(err)
	}
	dao.UpsertGame(game)
	dao.UpsertRating(rating)

	dao.ChangeScheduleResult(doubanId, consts.TypeGame, consts.ScheduleResultReady)
}

func processUser(doubanUid uint64) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r, " => ", util.GetCurrentGoroutineStack())
		}
	}()

	hash, _ := crawl.UserHash(doubanUid)
	rawUser := dao.GetUser(doubanUid)
	if rawUser != nil && rawUser.RssHash == hash {
		log.Println("user ", doubanUid, " not changed")
		return
	}
	//user
	user, err := crawl.UserOverview(doubanUid)
	if err != nil {
		dao.ChangeScheduleResult(doubanUid, consts.TypeUser, consts.ScheduleResultInvalid)
		panic(err)
	}
	dao.UpsertUser(user)

	//game
	if user.GameDo > 0 || user.GameWish > 0 || user.GameCollect > 0 {
		_, comment, game, err := crawl.CommentGame(doubanUid)
		if err != nil {
			panic(err)
		}
		for i, _ := range *game {
			added := dao.CreateGameNx(&(*game)[i])
			if added {
				dao.CreateSchedule((*game)[i].DoubanId, consts.TypeGame, consts.ScheduleStatusToCrawl, consts.ScheduleResultUnready)
				dao.UpsertComment(&(*comment)[i])
			}
		}
	}

	//
	if user.BookDo > 0 || user.BookWish > 0 || user.BookCollect > 0 {

		_, comment, book, err := crawl.CommentBook(doubanUid)
		if err != nil {
			panic(err)
		}

		for i, _ := range *book {
			added := dao.CreateBookNx(&(*book)[i])
			if added {
				dao.CreateSchedule((*book)[i].DoubanId, consts.TypeBook, consts.ScheduleStatusToCrawl, consts.ScheduleResultUnready)
				dao.UpsertComment(&(*comment)[i])

			}
		}
	}

	//movie
	if user.MovieDo > 0 || user.MovieWish > 0 || user.MovieCollect > 0 {

		_, comment, movie, err := crawl.CommentMovie(doubanUid)
		if err != nil {
			panic(err)
		}

		for i, _ := range *movie {
			added := dao.CreateMovieNx(&(*movie)[i])
			if added {
				dao.CreateSchedule((*movie)[i].DoubanId, consts.TypeMovie, consts.ScheduleStatusToCrawl, consts.ScheduleResultUnready)
				dao.UpsertComment(&(*comment)[i])
			}
		}
	}

	dao.ChangeScheduleResult(doubanUid, consts.TypeUser, consts.ScheduleResultReady)
}

func main() {
	ch := make(chan model.Schedule)

	for i := 1; i <= 5; i++ {
		go func(id int) {
			for {
				schedule := <-ch
				log.Println("agent ", id, " consume ", util.ToJson(schedule))
				switch schedule.Type {
				case consts.TypeBook:
					processBook(schedule.DoubanId)
					break
				case consts.TypeMovie:
					processMovie(schedule.DoubanId)
					break
				case consts.TypeGame:
					processGame(schedule.DoubanId)
					break
				case consts.TypeUser:
					processUser(schedule.DoubanId)
					break
				}
				dao.CasScheduleStatus(schedule.DoubanId, schedule.Type, consts.ScheduleStatusCrawled, consts.ScheduleStatusCrawling)
			}
		}(i)
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println(r, " => ", util.GetCurrentGoroutineStack())
			}
		}()
		lastIdle := false
		for {
			schedule := dao.SearchScheduleByStatus(consts.ScheduleStatusToCrawl)
			if schedule == nil {
				time.Sleep(time.Second * 5)
				if !lastIdle {
					log.Println("scanner idle")
				}
				lastIdle = true
			} else {
				lastIdle = false
				changed := dao.CasScheduleStatus(schedule.DoubanId, schedule.Type, consts.ScheduleStatusCrawling, consts.ScheduleStatusToCrawl)
				if changed {
					log.Println("scanner submit")
					ch <- *schedule
				}
			}
		}
	}()

	group := sync.WaitGroup{}
	group.Add(1)
	group.Wait()
}