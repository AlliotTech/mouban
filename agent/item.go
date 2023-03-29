package agent

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"mouban/consts"
	"mouban/dao"
	"mouban/model"
	"mouban/util"
	"strconv"
	"time"
)

func itemPendingSelector(t consts.Type, ch chan *model.Schedule) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorln(r, "item pending selector for", t.Name, "crashed  => ", util.GetCurrentGoroutineStack())
		}
	}()

	schedule := dao.SearchScheduleByStatus(t.Code, consts.ScheduleToCrawl.Code)
	if schedule != nil {
		ch <- schedule
	} else {
		time.Sleep(10 * time.Second)
	}
}

func itemRetrySelector(t consts.Type, ch chan *model.Schedule) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorln(r, "item retry selector for", t.Name, "crashed  => ", util.GetCurrentGoroutineStack())
		}
	}()

	schedule := dao.SearchScheduleByAll(t.Code, consts.ScheduleCrawled.Code, consts.ScheduleUnready.Code)
	if schedule != nil {
		ch <- schedule
	} else {
		time.Sleep(time.Minute)
	}
}

func itemDiscoverSelector(t consts.Type, ch chan *model.Schedule) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorln(r, "item discover selector for", t.Name, "crashed  => ", util.GetCurrentGoroutineStack())
		}
	}()

	schedule := dao.SearchScheduleByStatus(t.Code, consts.ScheduleCanCrawl.Code)

	if schedule != nil {
		ch <- schedule
	} else {
		time.Sleep(time.Minute)
	}
}

func itemWorker(index int, ch chan *model.Schedule) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorln(r, "item worker (", index, ") crashed  => ", util.GetCurrentGoroutineStack())
		}
	}()

	for schedule := range ch {
		t := consts.ParseType(schedule.Type)
		changed := dao.CasScheduleStatus(schedule.DoubanId, t.Code, consts.ScheduleCrawling.Code, *schedule.Status)
		if changed {
			logrus.Infoln("item thread", index, "start", t.Name, strconv.FormatUint(schedule.DoubanId, 10))
			processItem(schedule.Type, schedule.DoubanId)
			dao.CasScheduleStatus(schedule.DoubanId, t.Code, consts.ScheduleCrawled.Code, consts.ScheduleCrawling.Code)
			logrus.Infoln("item thread", index, "end", t.Name, strconv.FormatUint(schedule.DoubanId, 10))
		}
	}
}

func init() {
	if !viper.GetBool("agent.enable") {
		logrus.Infoln("item agent disabled")
		return
	}

	ch := make(chan *model.Schedule)

	types := []consts.Type{consts.TypeBook, consts.TypeMovie, consts.TypeGame, consts.TypeSong}
	for _, t := range types {
		go func() {
			for range time.NewTicker(time.Second).C {
				itemPendingSelector(t, ch)
			}
		}()
		go func() {
			for range time.NewTicker(time.Second).C {
				itemRetrySelector(t, ch)
			}
		}()
		go func() {
			for range time.NewTicker(time.Second).C {
				itemDiscoverSelector(t, ch)
			}
		}()
	}

	concurrency := viper.GetInt("agent.item.concurrency")
	for i := 0; i < concurrency; i++ {
		j := i + 1
		go func() {
			for range time.NewTicker(time.Second).C {
				itemWorker(j, ch)
			}
		}()
	}

	logrus.Infoln(concurrency, "item agent(s) enabled")
}
