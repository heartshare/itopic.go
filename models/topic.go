package models

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/russross/blackfriday"
)

//Topic struct
type Topic struct {
	TopicID      string
	Title        string
	Description  string
	Time         time.Time
	CategoryName string
	Content      string
	IsPublic     bool //true for public，false for protected
}

//MonthList Show The Topic Group By Month
type MonthList struct {
	Month  string
	Topics []*Topic
}

//Topics store all the topic
var Topics []*Topic

//TopicGroupByMonth store the topic by month
var TopicGroupByMonth []*MonthList

//InitTopicList Load All The Topic On Start
func InitTopicList() error {
	Topics = Topics[:0]
	return filepath.Walk("posts", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".md" {
			t, err := GetTopicByPath(path)
			if err != nil {
				return err
			}
			SetTopicToCategory(t)
			SetTopicToMonth(t)
			//按时间倒序排列
			for i := range Topics {
				if t.Time.After(Topics[i].Time) {
					Topics = append(Topics, nil)
					copy(Topics[i+1:], Topics[i:])
					Topics[i] = t
					return nil
				}
			}
			Topics = append(Topics, t)
		}
		return nil
	})
}

//GetTopicByPath Read The Topic By Path
func GetTopicByPath(path string) (*Topic, error) {
	fp, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	t := &Topic{
		Title:    strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
		IsPublic: true,
	}
	scanner := bufio.NewScanner(fp)

	for scanner.Scan() {
		s := scanner.Text()
		if strings.HasPrefix(s, "+++") {
			break
		}
		sp := strings.SplitN(s, ":", 2)
		if len(sp) != 2 {
			return nil, fmt.Errorf("invalid header: %s", s)
		}
		k, v := strings.TrimSpace(sp[0]), strings.TrimSpace(sp[1])
		switch k {
		case "url":
			t.TopicID = v
		case "des":
			t.Description = v
		case "time":
			t.Time, err = time.Parse("2006/01/02 15:04", v)
			if err != nil {
				return nil, err
			}
		case "category":
			t.CategoryName = v
		default:
			return nil, fmt.Errorf("invalid header: %s", s)
		}
	}
	var content bytes.Buffer
	for scanner.Scan() {
		content.Write(scanner.Bytes())
		content.WriteString("\n")
	}
	t.Content = string(blackfriday.MarkdownCommon(content.Bytes()))

	return t, nil
}

//SetTopicToCategory set topic to category struct
func SetTopicToCategory(t *Topic) {
	for k := range Categories {
		if Categories[k].CategoryID != t.CategoryName || t.IsPublic == false {
			continue
		}
		for i := range Categories[k].Topics {
			if t.Time.After(Categories[k].Topics[i].Time) {
				Categories[k].Topics = append(Categories[k].Topics, nil)
				copy(Categories[k].Topics[i+1:], Categories[k].Topics[i:])
				Categories[k].Topics[i] = t
				return
			}
		}
		Categories[k].Topics = append(Categories[k].Topics, t)
	}
}

//SetTopicToMonth set topic to month struct
func SetTopicToMonth(t *Topic) {
	month := t.Time.Format("2006-01")
	ml := &MonthList{}
	for _, m := range TopicGroupByMonth {
		if m.Month == month {
			ml = m
		}
	}
	if ml.Month == "" {
		ml.Month = month
		isFind := false
		for i := range TopicGroupByMonth {
			if strings.Compare(ml.Month, TopicGroupByMonth[i].Month) > 0 {
				TopicGroupByMonth = append(TopicGroupByMonth, nil)
				copy(TopicGroupByMonth[i+1:], TopicGroupByMonth[i:])
				TopicGroupByMonth[i] = ml
				isFind = true
				break
			}
		}
		if isFind == false {
			TopicGroupByMonth = append(TopicGroupByMonth, ml)
		}
	}
	for i := range ml.Topics {
		if t.Time.After(ml.Topics[i].Time) {
			ml.Topics = append(ml.Topics, nil)
			copy(ml.Topics[i+1:], ml.Topics[i:])
			ml.Topics[i] = t
			return
		}
	}
	ml.Topics = append(ml.Topics, t)
}