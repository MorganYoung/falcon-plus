// Copyright 2017 Xiaomi, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cron

import (
	"encoding/json"
	"fmt"
	"github.com/open-falcon/falcon-plus/modules/alarm/model"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

// RobotBody robot请求体
type RobotBody struct {
	Msgtype  string   `json:"msgtype"`
	Markdown Markdown `json:"markdown"`
}

// Markdown markdonw部分内容
type Markdown struct {
	MentionedList string `json:"mentioned_list"`
	Content       string `json:"content"`
}

func SendRobotList(L []*model.Robot) {
	for _, robot := range L {
		RobotWorkerChan <- 1
		go SendRobot(robot)
	}
}

func SendRobot(robot *model.Robot) {
	defer func() {
		<-RobotWorkerChan
	}()

	title := "falcon 邮件报警"
	urls := robot.Url
	if strings.Contains(urls, ",") {
		split := strings.Split(urls, ",")
		for _, url := range split {
			CallRobot(url, title, robot.Content)
		}
	} else {
		CallRobot(urls, title, robot.Content)
	}

}

func CallRobot(url string, title string, content string) {
	md := Markdown{"@all",
		fmt.Sprintf("# %s\n### %s", title, content)}
	body := RobotBody{"markdown", md}
	bodyStr, _ := json.Marshal(body)
	resp, err := http.Post(url, "application/json", strings.NewReader(string(bodyStr)))
	if err != nil {
		log.Errorf("send robot fail, url:%s, content:%s, error:%v", url, content, err)
	}

	log.Debugf("send robot:%v, resp:%v, url:%s", content, resp, url)
}
