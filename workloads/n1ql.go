package workloads

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type N1QL struct {
	Config       Config
	DeletedItems int64
	Default
}

func (w *N1QL) GenerateNewKey(currentRecords int64) string {
	return fmt.Sprintf("%012d", currentRecords)
}

func (w *N1QL) GenerateExistingKey(currentRecords int64) string {
	var randRecord int64
	total_records := currentRecords - w.DeletedItems
	hot_records := total_records * w.Config.HotDataPercentage / 100
	cold_records := total_records - hot_records
	if rand.Intn(100) < w.Config.HotSpotAccessPercentage {
		randRecord = 1 + w.DeletedItems + cold_records + rand.Int63n(hot_records)
	} else {
		randRecord = 1 + w.DeletedItems + rand.Int63n(cold_records)
	}
	return fmt.Sprintf("%012d", randRecord)
}

func (w *N1QL) GenerateKeyForRemoval() string {
	w.DeletedItems++
	return fmt.Sprintf("%012d", w.DeletedItems)
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func build_alphabet(key string) string {
	return Hash(key) + Hash(reverse(key))
}

func build_name(alphabet string) string {
	return fmt.Sprintf("%s %s", alphabet[:6], alphabet[6:12])
}

func build_email(alphabet string) string {
	return fmt.Sprintf("%s@%s.com", alphabet[12:18], alphabet[18:24])
}

func build_city(alphabet string) string {
	return alphabet[24:30]
}

func build_realm(alphabet string) string {
	return alphabet[30:36]
}

func build_country(alphabet string) string {
	return alphabet[42:48]
}

func build_county(alphabet string) string {
	return alphabet[48:54]
}

func build_street(alphabet string) string {
	return alphabet[54:62]
}

func build_coins(alphabet string) float64 {
	var coins, _ = strconv.ParseInt(alphabet[36:40], 16, 0)
	return math.Max(0.1, float64(coins)/100.0)
}

func build_category(alphabet string) int16 {
	var category, _ = strconv.ParseInt(string(alphabet[41]), 16, 0)
	return int16(category % 3)
}

func build_year(alphabet string) int16 {
	var year, _ = strconv.ParseInt(string(alphabet[62]), 32, 0)
	return int16(1985 + year)
}

func build_state(alphabet string) string {
	idx := strings.Index(alphabet, "7") % NUM_STATES
	if idx == -1 {
		idx = 56
	}
	return STATES[idx][0]
}

func build_full_state(alphabet string) string {
	idx := strings.Index(alphabet, "8") % NUM_STATES
	if idx == -1 {
		idx = 56
	}
	return STATES[idx][1]
}

func build_gmtime(alphabet string) []int16 {
	var id, _ = strconv.ParseInt(string(alphabet[63]), 16, 0)
	seconds := 396 * 24 * 3600 * (id % 12)
	d := time.Duration(seconds) * time.Second
	t := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC).Add(d)

	return []int16{
		int16(t.Year()),
		int16(t.Month()),
		int16(t.Day()),
		int16(t.Hour()),
		int16(t.Minute()),
		int16(t.Second()),
		int16(t.Weekday() - 1),
		int16(t.YearDay()),
		int16(0),
	}
}

func build_achievements(alphabet string) (achievements []int16) {
	achievement := int16(256)
	for i, char := range alphabet[42:58] {
		var id, _ = strconv.ParseInt(string(char), 16, 0)
		achievement = (achievement + int16(id)*int16(i)) % 512
		if achievement < 256 {
			achievements = append(achievements, achievement)
		}
	}
	return
}

func (w *N1QL) GenerateValue(key string, size int) map[string]interface{} {
	alphabet := build_alphabet(key)

	return map[string]interface{}{
		"name":         build_name(alphabet),
		"email":        build_email(alphabet),
		"city":         build_city(alphabet),
		"realm":        build_realm(alphabet),
		"country":      build_country(alphabet),
		"county":       build_county(alphabet),
		"street":       build_street(alphabet),
		"coins":        build_coins(alphabet),
		"year":         build_year(alphabet),
		"category":     build_category(alphabet),
		"state":        build_state(alphabet),
		"full_state":   build_full_state(alphabet),
		"achievements": build_achievements(alphabet),
		"gmtime":       build_gmtime(alphabet),
	}
}