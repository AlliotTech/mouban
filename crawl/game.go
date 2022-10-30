package crawl

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	"mouban/consts"
	"mouban/model"
	"strings"
)

func Game(doubanId uint64) (*model.Game, *model.Rating, error) {
	body, err := Get(fmt.Sprintf(consts.GameDetailUrl, doubanId))
	if err != nil {
		return nil, nil, err
	}

	doc, err := htmlquery.Parse(strings.NewReader(*body))
	if err != nil {
		return nil, nil, err
	}
	list := htmlquery.Find(doc, "//a[@href]")
	for i := range list {
		fmt.Println(list[i])
	}

	return nil, nil, nil
}
