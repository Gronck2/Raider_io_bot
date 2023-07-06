package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	URL         = "https://raider.io"
	mythicScore = "%s/api/v1/characters/profile"
	affixes     = "%s/api/v1/mythic-plus/affixes"
)

func GetScore(region, name, realm string) (*Info, error) {
	emptyErr := errors.New("")
	client := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf(mythicScore, URL), nil)

	q := req.URL.Query()
	q.Add("region", region)
	q.Add("name", name)
	q.Add("realm", realm)
	q.Add("fields", "mythic_plus_scores_by_season:current,gear,guild")

	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)

	if err != nil {
		return nil, emptyErr
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, emptyErr
	}
	fmt.Println(resp.StatusCode)
	if resp.StatusCode != 200 {
		e := &ApiError{}
		err = json.Unmarshal(body, e)
		if err != nil {
			return nil, emptyErr
		}
		log.Println(e.Message)
		return nil, errors.New(e.Message)
	}

	info := &Info{}
	err = json.Unmarshal(body, info)
	if err != nil {
		return nil, emptyErr
	}

	return info, nil
}

func GetAffixes(region, locale string) (*Affixes, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf(affixes, URL), nil)

	q := req.URL.Query()
	q.Add("region", region)
	q.Add("locale", locale)

	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	affixes := &Affixes{}
	err = json.Unmarshal(body, affixes)
	if err != nil {
		return nil, err
	}

	return affixes, nil
}
