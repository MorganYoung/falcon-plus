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

package api

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/modules/alarm/g"
	"github.com/open-falcon/falcon-plus/modules/api/app/model/uic"
	"github.com/toolkits/container/set"
	"github.com/toolkits/net/httplib"
	"strings"
	"sync"
	"time"
)

type APIGetTeamOutput struct {
	Team        uic.Team    `json:"team"`
	Users       []*uic.User `json:"users"`
	TeamCreator string      `json:"creator_name"`
}

type UsersCache struct {
	sync.RWMutex
	M map[string][]*uic.User
}

type TeamsCache struct {
	sync.RWMutex
	M map[string]uic.Team
}

var Users = &UsersCache{M: make(map[string][]*uic.User)}

var Teams = &TeamsCache{M: make(map[string]uic.Team)}

func (this *UsersCache) Get(team string) []*uic.User {
	this.RLock()
	defer this.RUnlock()
	val, exists := this.M[team]
	if !exists {
		return nil
	}

	return val
}

func (this *UsersCache) Set(team string, users []*uic.User) {
	this.Lock()
	defer this.Unlock()
	this.M[team] = users
}

func (this *TeamsCache) Get(team string) uic.Team {
	this.RLock()
	defer this.RUnlock()
	val, exists := this.M[team]
	if !exists {
		return uic.Team{}
	}

	return val
}

func (this *TeamsCache) Set(team string, teamr uic.Team) {
	this.Lock()
	defer this.Unlock()
	this.M[team] = teamr
}

func UsersOf(team string) []*uic.User {
	users := CurlUic(team)

	if users != nil {
		Users.Set(team, users)
	} else {
		users = Users.Get(team)
	}

	return users
}

func TeamOf(team string) uic.Team {
	teamr := CurlTeam(team)

	if teamr.Name != "" {
		Teams.Set(team, teamr)
	} else {
		teamr = Teams.Get(team)
	}

	return teamr
}

func GetUsers(teams string) map[string]*uic.User {
	userMap := make(map[string]*uic.User)
	arr := strings.Split(teams, ",")
	for _, team := range arr {
		if team == "" {
			continue
		}

		users := UsersOf(team)
		if users == nil {
			continue
		}

		for _, user := range users {
			userMap[user.Name] = user
		}
	}
	return userMap
}

func GetTeams(teams string) map[string]*uic.Team {
	teamMap := make(map[string]*uic.Team)
	arr := strings.Split(teams, ",")
	for _, team := range arr {
		if team == "" {
			continue
		}

		teamr := TeamOf(team)
		if teamr.Name == "" {
			continue
		}
		teamMap[team] = &teamr

	}
	return teamMap
}

// return phones, emails, IM
func ParseTeams(teams string) ([]string, []string, []string, []string) {
	if teams == "" {
		return []string{}, []string{}, []string{}, []string{}
	}

	userMap := GetUsers(teams)
	phoneSet := set.NewStringSet()
	mailSet := set.NewStringSet()
	imSet := set.NewStringSet()
	for _, user := range userMap {
		if user.Phone != "" {
			phoneSet.Add(user.Phone)
		}
		if user.Email != "" {
			mailSet.Add(user.Email)
		}
		if user.IM != "" {
			imSet.Add(user.IM)
		}
	}

	teamMap := GetTeams(teams)
	robotSet := set.NewStringSet();
	for _, team := range teamMap {
		if team.Robot != "" {
			robotSet.Add(team.Robot)
		}
	}

	return phoneSet.ToSlice(), mailSet.ToSlice(), imSet.ToSlice(), robotSet.ToSlice()
}

func CurlUic(team string) []*uic.User {
	if team == "" {
		return []*uic.User{}
	}

	uri := fmt.Sprintf("%s/api/v1/team/name/%s", g.Config().Api.PlusApi, team)
	req := httplib.Get(uri).SetTimeout(2*time.Second, 10*time.Second)
	token, _ := json.Marshal(map[string]string{
		"name": "falcon-alarm",
		"sig":  g.Config().Api.PlusApiToken,
	})
	req.Header("Apitoken", string(token))

	var team_users APIGetTeamOutput
	err := req.ToJson(&team_users)
	if err != nil {
		log.Errorf("curl %s fail: %v", uri, err)
		return nil
	}

	return team_users.Users
}

func CurlTeam(team string) uic.Team {
	if team == "" {
		return uic.Team{}
	}

	uri := fmt.Sprintf("%s/api/v1/team/name/%s", g.Config().Api.PlusApi, team)
	req := httplib.Get(uri).SetTimeout(2*time.Second, 10*time.Second)
	token, _ := json.Marshal(map[string]string{
		"name": "falcon-alarm",
		"sig":  g.Config().Api.PlusApiToken,
	})
	req.Header("Apitoken", string(token))

	var team_users APIGetTeamOutput
	err := req.ToJson(&team_users)
	if err != nil {
		log.Errorf("curl %s fail: %v", uri, err)
		return uic.Team{}
	}

	return team_users.Team
}
