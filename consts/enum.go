package consts

type RatingStatus struct {
	Code uint8
	Name string
}

var RatingNormal = RatingStatus{0, "normal"}
var RatingNotEnough = RatingStatus{1, "not enough"}
var RatingNotAllowed = RatingStatus{2, "not allowed"}

type Action struct {
	Code uint8
	Name string
}

var ActionDo = Action{0, "do"}
var ActionWish = Action{1, "wish"}
var ActionCollect = Action{2, "collect"}
var ActionHide = Action{3, "hide"}

type Type struct {
	Code uint8
	Name string
}

var TypeUser = Type{0, "user"}
var TypeBook = Type{1, "book"}
var TypeMovie = Type{2, "movie"}
var TypeGame = Type{3, "game"}
var TypeSong = Type{4, "song"}

func ParseType(code uint8) Type {
	switch code {
	case 0:
		return TypeUser
	case 1:
		return TypeBook
	case 2:
		return TypeMovie
	case 3:
		return TypeGame
	case 4:
		return TypeSong
	default:
		return TypeUser
	}
}

type ScheduleStatus struct {
	Code uint8
	Name string
}

var ScheduleToCrawl = ScheduleStatus{0, "to crawl"}
var ScheduleCrawling = ScheduleStatus{1, "crawling"}
var ScheduleCrawled = ScheduleStatus{2, "crawled"}
var ScheduleCanCrawl = ScheduleStatus{3, "can crawl"}

type ScheduleResult struct {
	Code uint8
	Name string
}

var ScheduleUnready = ScheduleResult{0, "unready"}
var ScheduleReady = ScheduleResult{1, "ready"}
var ScheduleInvalid = ScheduleResult{2, "invalid"}

func ParseResult(code uint8) ScheduleResult {
	switch code {
	case 0:
		return ScheduleUnready
	case 1:
		return ScheduleReady
	case 2:
		return ScheduleInvalid
	default:
		return ScheduleUnready
	}
}
