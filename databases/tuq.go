package databases

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
)

type RestClient struct {
	client *http.Client
	URIs   []string
}

func (c RestClient) Do(q string) error {
	data := bytes.NewReader([]byte(q))
	uri := c.URIs[rand.Intn(len(c.URIs))]
	req, err := http.NewRequest("POST", uri, data)
	req.Header.Set("Content-Type", "text/plain")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("Bad status code")
	}

	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)

	return err
}

type Tuq struct {
	client *RestClient
	cb     Couchbase
	bucket string
}

const MaxIdleConnsPerHost = 1000

func (t *Tuq) Init(config Config) {
	URIs := []string{}
	for _, address := range config.Addresses {
		URIs = append(URIs, fmt.Sprintf("%squery", address))
	}

	tr := &http.Transport{MaxIdleConnsPerHost: MaxIdleConnsPerHost}
	t.client = &RestClient{&http.Client{Transport: tr}, URIs}
	t.bucket = config.Table

	t.cb = Couchbase{}
	t.cb.Init(config)
}

func (t *Tuq) Shutdown() {}

func (t *Tuq) Create(key string, value map[string]interface{}) error {
	return t.cb.Create(key, value)
}

func (t *Tuq) Read(key string) error {
	return t.cb.Read(key)
}

func (t *Tuq) Update(key string, value map[string]interface{}) error {
	return t.cb.Update(key, value)
}

func (t *Tuq) Delete(key string) error {
	return t.cb.Delete(key)
}

func (t *Tuq) Query(key string, args []interface{}) error {
	index := args[0].(string)

	var q string
	switch index {
	case "name_and_street_by_city":
		query := `
			SELECT name.f.f.f AS _name, street.f.f AS _street
				FROM %s
				WHERE city.f.f = "%s"
				LIMIT 20`
		q = fmt.Sprintf(query, t.bucket, args[1])
	case "name_and_email_by_county":
		query := `
			SELECT name.f.f.f AS _name, email.f.f AS _email
				FROM %s
				WHERE county.f.f = "%s"
				LIMIT 20`
		q = fmt.Sprintf(query, t.bucket, args[1])
	case "achievements_by_realm":
		query := `
			SELECT achievements
				FROM %s
				WHERE realm.f = "%s"
				LIMIT 20`
		q = fmt.Sprintf(query, t.bucket, args[1])
	case "name_by_coins":
		query := `
			SELECT name.f.f.f AS _name
				FROM %s
				WHERE coins.f > %f AND coins.f < %f
				LIMIT 20`
		q = fmt.Sprintf(query, t.bucket, args[1].(float64)*0.5, args[1])
	case "email_by_achievement_and_category":
		query := `
			SELECT email.f.f AS _email
				FROM %s
				WHERE category = %d AND achievements[0] > 0 AND achievements[0] < %d
				LIMIT 20`
		q = fmt.Sprintf(query, t.bucket, args[2].(int16), args[1].([]int16)[0])
	case "street_by_year_and_coins":
		query := `
			SELECT street.f.f AS _street
				FROM %s
				WHERE year = %d AND coins.f > %f AND coins.f < 655.35
				LIMIT 20`
		q = fmt.Sprintf(query, t.bucket, args[1], args[2].(float64))
	case "name_and_email_and_street_and_achievements_and_coins_by_city":
		query := `
			SELECT name.f.f.f AS _name, email.f.f AS _email, street.f.f AS _street, achievements, coins.f AS _coins
				FROM %s
				WHERE city.f.f = "%s"
				LIMIT 20`
		q = fmt.Sprintf(query, t.bucket, args[1])
	case "street_and_name_and_email_and_achievement_and_coins_by_county":
		query := `
			SELECT street.f.f AS _street, name.f.f.f AS _name, email.f.f AS _email, achievements[0] AS achievement, 2*coins.f AS _coins
				FROM %s
				WHERE county.f.f = "%s"
				LIMIT 20`
		q = fmt.Sprintf(query, t.bucket, args[1])
	case "category_name_and_email_and_street_and_gmtime_and_year_by_country":
		query := `
			SELECT category, name.f.f.f AS _name, email.f.f AS _email, street.f.f AS _street, gmtime, year
				FROM %s
				WHERE country.f = "%s"
				LIMIT 20`
		q = fmt.Sprintf(query, t.bucket, args[1])
	case "body_by_city":
		query := `
			SELECT body
				FROM %s
				WHERE city.f.f = "%s"
				LIMIT 20`
		q = fmt.Sprintf(query, t.bucket, args[1])
	case "body_by_realm":
		query := `
			SELECT body
				FROM %s
				WHERE realm.f = "%s"
				LIMIT 20`
		q = fmt.Sprintf(query, t.bucket, args[1])
	case "body_by_country":
		query := `
			SELECT body
				FROM %s
				WHERE country.f = "%s"
				LIMIT 20`
		q = fmt.Sprintf(query, t.bucket, args[1])
	case "distinct_states":
		query := `
			SELECT DISTINCT state.f AS state
				FROM %s
				LIMIT 20`
		q = fmt.Sprintf(query, t.bucket)
	case "distinct_full_states":
		query := `
			SELECT DISTINCT full_state.f AS full_state
				FROM %s
				LIMIT 20`
		q = fmt.Sprintf(query, t.bucket)
	case "distinct_years":
		query := `
			SELECT DISTINCT year
				FROM %s
				LIMIT 20`
		q = fmt.Sprintf(query, t.bucket)
	case "coins_stats_by_state_and_year":
		query := `
			SELECT COUNT(coins.f), SUM(coins.f), AVG(coins.f), MIN(coins.f), MAX(coins.f)
				FROM %s
				WHERE state.f = "%s" AND year = %d
				GROUP BY state.f, year
				LIMIT 20`
		q = fmt.Sprintf(query, t.bucket, args[1], args[2])
	case "coins_stats_by_gmtime_and_year":
		query := `
			SELECT COUNT(coins.f), SUM(coins.f), AVG(coins.f), MIN(coins.f), MAX(coins.f)
				FROM %s
				WHERE gmtime = [%d, %d, %d, %d, %d, %d, %d, %d, %d] AND year = %d
				GROUP BY gmtime, year
				LIMIT 20`
		gmtime := args[1].([]int16)
		q = fmt.Sprintf(query, t.bucket,
			gmtime[0], gmtime[1], gmtime[2], gmtime[3], gmtime[4], gmtime[5], gmtime[6], gmtime[7], gmtime[8],
			args[2])
	case "coins_stats_by_full_state_and_year":
		query := `
			SELECT COUNT(coins.f), SUM(coins.f), AVG(coins.f), MIN(coins.f), MAX(coins.f)
				FROM %s
				WHERE full_state.f = "%s" AND year = %d
				GROUP BY full_state.f, year
				LIMIT 20`
		q = fmt.Sprintf(query, t.bucket, args[1], args[2])
	}

	return t.client.Do(q)
}
