package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	baseURL = "http://test.hnyuntong.cn/api/v2/message/send?conn=wifi&icc=&ua=OnePlusKB2000&ram=7894007808&cc=TG85595&dev_name=OnePlus&ndid=&ver=dev&cv=LOVEMOONS3.5.40_Android"
	content = `{"content": {
		"audio_content": {
			"duration": 0,
			"localUrl": "",
			"url": ""
		},
		"extra_comment": {
			"highlight_list": [],
			"text_list": []
		},
		"finan_attr": {
			"earn": 0,
			"earn_text": "",
			"need_show_tips": 0,
			"tips": ""
		},
		"gift_content": {
			"combo_code": "",
			"gift_id": -1,
			"gift_name": "",
			"gold": -1,
			"intimacy_value": -1.0,
			"res_id": -1,
			"seq": -1,
			"sub_res": {
				"bundle": 1,
				"lucky_gold": -1,
				"lucky_id": -1,
				"lucky_name": ""
			},
			"gift_type": -1
		},
		"guard_card_content": {
			"guard_hour": "",
			"invite_expire": 0,
			"status": 0,
			"text": ""
		},
		"highlight": [],
		"image_content": {
			"height": 0,
			"localUrl": "",
			"url": "",
			"width": 0
		},
		"links": [],
		"love_letter_content": {
			"content": "",
			"effect_id": "",
			"letter_id": 0,
			"love_letter_stencil_info": {
				"button_img": "",
				"gold": 0,
				"icon": "",
				"img": "",
				"limit_num": 0,
				"stencil_id": 0,
				"text_color": "",
				"thumbnail": "",
				"title": ""
			},
			"recv_nick": "",
			"recv_uid": 0,
			"send_nick": "",
			"send_uid": 0,
			"time": 0
		},
		"push_jump_content": {
			"content": "",
			"content_highlights": [],
			"jump_text": "",
			"jump_text_highlights": [],
			"jump_url": "",
			"pic": "",
			"title": "",
			"title_highlights": []
		},
		"text_content": {
			"content": "情书定制",
			"highlight_content": "",
			"highlights": [],
			"leading": [],
			"url": ""
		},
		"tips": "",
		"tips_img": ""
	},
	"logicVersion": 3,
	"peer_id": 100295,
	"seq_id": 1636612972458,
	"type": 1}`
)

func main() {
	opt, err := redis.ParseURL("redis://10.100.130.15:6379/2")
	if err != nil {
		panic(err)
	}

	opt.Password = "tSbe0mOxh5qq"
	r := redis.NewClient(opt)
	c := http.Client{}

	u, _ := url.Parse(baseURL)

	var keys []string
	var cursor uint64
	for {
		keys, cursor, err = r.ZScan("hall_popular_1", cursor, "*", 100).Result()
		if err != nil {
			panic(err)
		}
		for i := 0; i < len(keys); i += 2 {
			q := u.Query()
			q.Set("uid", keys[i])
			u.RawQuery = q.Encode()

			rsp, err := c.Post(u.String(), "application/json", strings.NewReader(content))
			if err != nil {
				panic(err)
			}

			result, err := ioutil.ReadAll(rsp.Body)
			if err != nil {
				panic(err)
			}
			rsp.Body.Close()

			fmt.Printf("respond: %s\n", string(result))

			if cursor == 0 {
				break
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}
