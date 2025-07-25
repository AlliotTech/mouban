package crawl

import (
	"errors"
	"fmt"
	"mouban/consts"
	"mouban/model"
	"mouban/util"
	"strconv"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

func Book(doubanId uint64) (*model.Book, *model.Rating, *[]string, *[]uint64, error) {
	body, _, err := Get(fmt.Sprintf(consts.BookDetailUrl, doubanId), ItemLimiter)
	if err != nil {
		if strings.Contains(err.Error(), "too many redirects") {
			return nil, nil, nil, nil, err
		}
		panic(err)
	}

	doc, err := htmlquery.Parse(strings.NewReader(*body))

	if err != nil {
		panic(err)
	}

	tt := htmlquery.FindOne(doc, "//head//title")
	if tt == nil {
		panic("title is nil for " + strconv.FormatUint(doubanId, 10) + ", html: " + htmlquery.OutputHTML(doc, true))
	}
	t := htmlquery.InnerText(tt)
	if strings.TrimSpace(t) == "页面不存在" || strings.TrimSpace(t) == "条目不存在" {
		return nil, nil, nil, nil, errors.New(strings.TrimSpace(t))
	}

	ttt := htmlquery.FindOne(doc, "//div[@id='exception']")
	if ttt != nil {
		exceptionInfo := htmlquery.InnerText(ttt)
		if strings.Contains(exceptionInfo, "根据相关法律法规") {
			return nil, nil, nil, nil, errors.New(exceptionInfo)
		}
	}

	title := htmlquery.SelectAttr(htmlquery.FindOne(doc, "//meta[@property='og:title']"), "content")
	thumbnail := htmlquery.SelectAttr(htmlquery.FindOne(doc, "//a[@class='nbg']/img"), "src")
	intros := htmlquery.Find(doc, "//div[@class='intro']")
	var selected []*html.Node
	for _, intro := range intros {
		if strings.Contains(htmlquery.InnerText(intro), "(展开全部)") {
			continue
		}
		selected = append(selected, intro)
	}

	contentIntro := ""
	if len(selected) >= 1 {
		contentIntro = util.TrimBookParagraph(selected[0])
	}

	authorIntro := ""
	if len(selected) >= 2 {
		authorIntro = util.TrimBookParagraph(selected[1])
	}

	info := htmlquery.FindOne(doc, "//div[@id='info']")
	if info == nil {
		panic("info is nil for " + strconv.FormatUint(doubanId, 10) + ", html: " + htmlquery.OutputHTML(doc, true))
	}
	data := util.TrimInfo(htmlquery.OutputHTML(info, false))

	isbn := strings.TrimSpace(data["ISBN"])
	subtitle := strings.TrimSpace(data["副标题"])
	orititle := strings.TrimSpace(data["原作名"])
	author := strings.TrimSpace(data["作者"])
	press := strings.TrimSpace(data["出版社"])
	producer := strings.TrimSpace(data["出品方"])
	translator := strings.TrimSpace(data["译者"])
	serial := strings.TrimSpace(data["丛书"])
	publishDate := strings.TrimSpace(data["出版年"])
	framing := strings.TrimSpace(data["装帧"])
	page := uint32(util.ParseNumber(data["页数"]))
	price := uint32(util.ParseFloat(data["定价"]) * 100)

	book := &model.Book{
		DoubanId:    doubanId,
		Title:       title,
		Subtitle:    subtitle,
		Orititle:    orititle,
		Author:      author,
		Translator:  translator,
		Press:       press,
		Producer:    producer,
		Serial:      serial,
		PublishDate: publishDate,
		ISBN:        isbn,
		Framing:     framing,
		Page:        page,
		Price:       price,
		BookIntro:   contentIntro,
		AuthorIntro: authorIntro,
		Thumbnail:   thumbnail,
	}

	rating := Rating(htmlquery.FindOne(doc, "//div[@id='interest_sectl']"))
	rating.DoubanId = doubanId
	rating.Type = consts.TypeBook.Code

	newUsers := util.ParseNewUsers(doc)
	newItems := util.ParseNewItems(doc, consts.TypeBook)

	return book, rating, newUsers, newItems, nil
}
