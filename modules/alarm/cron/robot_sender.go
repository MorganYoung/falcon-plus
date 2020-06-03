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
	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/modules/alarm/g"
	"github.com/open-falcon/falcon-plus/modules/alarm/model"
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

func SendRobotList(L []*model.Mail) {
	for _, mail := range L {
		RobotWorkerChan <- 1
		go SendRobot(mail)
	}
}

func SendRobot(mail *model.Mail) {
	defer func() {
		<-RobotWorkerChan
	}()

	url := g.Config().Api.ROBOT
	md := Markdown{"@all",
		fmt.Sprintf("# %s\n### %s","falcon 邮件报警", mail.Content)}
	body := RobotBody{"markdown", md}
	bodyStr, _ := json.Marshal(body)
	resp, err := http.Post(url, "application/json", strings.NewReader(string(bodyStr)))
	if err != nil {
		log.Errorf("send im fail, tos:%s, content:%s, error:%v", mail.Tos, mail.Content, err)
	}

	log.Debugf("send im:%v, resp:%v, url:%s", mail, resp, url)
}
