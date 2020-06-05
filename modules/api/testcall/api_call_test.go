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

package testcall

import (
	"encoding/json"
	"fmt"
	"github.com/masato25/resty"
	"github.com/open-falcon/falcon-plus/modules/api/app/model/uic"
	cfg "github.com/open-falcon/falcon-plus/modules/api/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"testing"
)

var (
	api_v1             = ""
	test_user_name     = "apitest-user1"
	test_user_password = "password"
	test_team_name     = "apitest-team1"
	root_user_name     = "morgan"
	root_user_password = "morgan"
	api_token          = ""
)

func init() {
	cfg_file := os.Getenv("API_TEST_CFG")
	if cfg_file == "" {
		cfg_file = "./cfg.example"
	}
	viper.SetConfigName(cfg_file)
	viper.AddConfigPath(".")
	viper.AddConfigPath("../")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}

	err = cfg.InitLog(viper.GetString("log_level"))
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatalf("db conn failed with error %s", err.Error())
	}

	api_port := 8070

	api_host := os.Getenv("API_HOST")
	if api_host == "" {
		api_host = "127.0.0.1"
	}
	api_v1 = fmt.Sprintf("http://%s:%s/api/v1", api_host, api_port)
	fmt.Println(api_v1)
}

func TestTeamApi(t *testing.T) {
	var rr *map[string]interface{} = &map[string]interface{}{}
	resp1, _ := resty.R().
		SetQueryParam("name", root_user_name).SetQueryParam("password", root_user_password).SetResult(rr).
		Post(fmt.Sprintf("%s/user/login", api_v1))
	log.Info(resp1)
	api_token = fmt.Sprintf(`{"name": "%v", "sig": "%v"}`, (*rr)["name"], (*rr)["sig"])

	*rr = map[string]interface{}{}
	resp, _ := resty.R().
		SetHeader("Apitoken", api_token).
		SetResult(rr).
		Get(fmt.Sprintf("%s/team/t/%v", api_v1, 2))
	log.Println(resp)

	team := "team2"
	uri := fmt.Sprintf("%s/team/name/%s", api_v1, team)
	resp, _ = resty.R().
		SetHeader("Apitoken", api_token).
		SetResult(rr).
		Get(uri)
	//req := httplib.Get(uri).SetTimeout(2*time.Second, 10*time.Second)

	var team_users APIGetTeamOutput
	err := json.Unmarshal(resp.Body(), &team_users)
	if err != nil {
		log.Errorf("curl %s fail: %v", uri, err)
	}
	log.Println(team_users)
	log.Println(team_users.Team)
	tamJson, _ := json.Marshal(team_users.Team)
	log.Println(string(tamJson))
}

type APIGetTeamOutput struct {
	uic.Team
	Users       []*uic.User `json:"users"`
	TeamCreator string      `json:"creator_name"`
}
