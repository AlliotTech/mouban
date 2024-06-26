package crawl

import (
	"mouban/consts"
	"mouban/model"
	"mouban/util"
	"strconv"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

func Rating(interestSelect *html.Node) *model.Rating {
	if interestSelect == nil {
		result := &model.Rating{
			Status: &consts.RatingNotAllowed.Code,
		}
		return result
	}

	ratingRaw := htmlquery.InnerText(htmlquery.FindOne(interestSelect, "//strong[@property='v:average']"))
	if len(strings.TrimSpace(ratingRaw)) == 0 {
		result := &model.Rating{
			Status: &consts.RatingNotEnough.Code,
		}
		return result
	}
	rating := util.ParseFloat(ratingRaw)
	totalStr := htmlquery.InnerText(htmlquery.FindOne(interestSelect, "//span[@property='v:votes']"))
	total, err := strconv.ParseUint(totalStr, 10, 32)
	if err != nil {
		return nil
	}
	stars := htmlquery.Find(interestSelect, "//span[@class='rating_per']")

	star5Str := strings.TrimSpace(htmlquery.InnerText(stars[0]))
	star5, _ := strconv.ParseFloat(star5Str[0:len(star5Str)-1], 32)
	star4Str := strings.TrimSpace(htmlquery.InnerText(stars[1]))
	star4, _ := strconv.ParseFloat(star4Str[0:len(star4Str)-1], 32)
	star3Str := strings.TrimSpace(htmlquery.InnerText(stars[2]))
	star3, _ := strconv.ParseFloat(star3Str[0:len(star3Str)-1], 32)
	star2Str := strings.TrimSpace(htmlquery.InnerText(stars[3]))
	star2, _ := strconv.ParseFloat(star2Str[0:len(star2Str)-1], 32)
	star1Str := strings.TrimSpace(htmlquery.InnerText(stars[4]))
	star1, _ := strconv.ParseFloat(star1Str[0:len(star1Str)-1], 32)

	result := &model.Rating{
		Total:  uint32(total),
		Rating: rating,
		Star5:  float32(star5),
		Star4:  float32(star4),
		Star3:  float32(star3),
		Star2:  float32(star2),
		Star1:  float32(star1),
		Status: &consts.RatingNormal.Code,
	}
	return result
}
