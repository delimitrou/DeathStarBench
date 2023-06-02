package main

import (
    "fmt"
    "time"
    "sort"
    "math/rand"
    "math"
    "log"
    "os"
)

// data structures for video MetaData
type Meta struct {
	// user that uploads the video
	UserId string `json:"user_id"`
	// resolution of the video
	Resolution string `json:"resolution"`
	// duration of the video
	Duration float64 `json:"duration"`
	// text description of the video
	Description string `json:"description"`
	// date that this video is uploaded
	Date string `json:"date"`
}
type Rating struct {
	// number of ratings
	Num int64 `json:"num"`
	// mean score of all ratings
	Score float64 `json:"score"`
	// mean squared score of all ratings
	ScoreSq float64 `json:"score_sq"`
}
// data structre for video info
type Info struct {
	VideoMeta Meta  `json:"meta"`
	// number of views
	Views int64 `json:"views"`
	Rate Rating  `json:"rating"`
}

var logger = log.New(os.Stdout, "", 0)
var dateLayout string = "2006-01-02"
func TimeToDate(t time.Time) string {
	return t.Format(dateLayout)
}
func DateToTime(d string) (time.Time, error) {
	return time.Parse(dateLayout, d)
}
// DatesBetween return all the dates between the given start & end date
// an error is generated is startDate or endDate does not conform to dateLayout
func DatesBetween(startDate string, endDate string) ([]string, error) {
	startT, err := DateToTime(startDate)
	if err != nil {
		return nil, err
	}
	endT, err := DateToTime(endDate)
	if err != nil {
		return nil, err
	}
	dates := make([]string, 0)
	for d := startT; !d.After(endT); d = d.AddDate(0, 0, 1) {
		dates = append(dates, TimeToDate(d))
	}
	return dates, nil
}
// DaysBetween return the number of days between two given dates
func DaysBetween(start time.Time, end time.Time) int {
	if start.After(end) {
        start, end = end, start
    }
    return int(end.Sub(start).Hours() / 24)
}

// TrendingScore is used to sort videos by trending
type TrendingScore struct {
	VideoId string
	Score float64
	Date time.Time
}
// normalizing range used to rank videos
var (
	upperView float64 = 3.0
	upperRate float64 = 1.0
)
// return the decayed score based on freshness
func freshDecay(score float64, days int) float64 {
	return score*math.Exp(-float64(days)/30.0)
}
func sortVideos(videoinfo map[string]Info) ([]TrendingScore, error) {
	// normalize view score to [0, upperView] and rate score to [0, upperRate]
	// rate score is computed with lower 68% (around delta) confid interval 
	// view score
	var maxviews int64 = 0
	var minviews int64 = 0
	var maxrate float64 = 0.0
	var minrate float64 = 0.0
	// used to compute delta of the entire dataset
	var totalscore float64 = 0
	var totalscoresq float64 = 0
	var totalnum int64 = 0
	views := make(map[string]int64)
	var firstv bool = true
	rates := make(map[string]float64)
	var firstr bool = true
	undecided := make([]string, 0)
	dates := make(map[string]string)
	for vid, vinfo := range videoinfo {
		if firstv || maxviews < vinfo.Views {
			maxviews = vinfo.Views
		}
		if firstv || minviews > vinfo.Views {
			minviews = vinfo.Views
		}
		firstv = false
		views[vid] = vinfo.Views
		n := vinfo.Rate.Num
		s := vinfo.Rate.Score
		ssq := vinfo.Rate.ScoreSq
		totalnum += n
		totalscore += s*float64(n)
		totalscoresq += ssq*float64(n)
		var delta float64 = ssq - math.Pow(s, 2)
		if delta < 0 {
			logger.Printf("Error: delta of video:%s < 0 (score_sq = %.3f, score = %.3f)",
				vid, ssq, s)
			err := fmt.Errorf("delta of video:%s < 0 (score_sq = %.3f, score = %.3f)",
				vid, ssq, s)
			return nil, err
		} 
		if delta == 0.0 || n == 0 {
			rates[vid] = s
			undecided = append(undecided, vid)
		} else {
			rates[vid] = s - delta/math.Sqrt(float64(n))
			if firstr || maxrate < rates[vid] {
				maxrate = rates[vid]
			}
			if firstr || minrate > rates[vid] {
				minrate = rates[vid]
			}
			firstr = false
		}
		dates[vid] = vinfo.VideoMeta.Date
	}
	// compute undecided
	totaldelta := totalscoresq/float64(totalnum) - math.Pow(totalscore/float64(totalnum), 2)
	for _, vid := range undecided {
		rates[vid] -= totaldelta
		if firstr || maxrate < rates[vid] {
			maxrate = rates[vid]
		}
		if firstr || minrate > rates[vid] {
			minrate = rates[vid]
		}
		firstr = false 
	}
	// normalize scores and scale all scores based on freshness
	viewrange := maxviews - minviews
	if viewrange == 0 {
		viewrange = 1
	}
	raterange := maxrate - minrate
	if raterange == 0.0 {
		raterange = 1.0
	}
	now := time.Now()
	trendings := make([]TrendingScore, 0)
	for vid, _ := range videoinfo {
		viewscore := float64(views[vid] - minviews) / float64(viewrange) * upperView
		ratescore := (rates[vid] - minrate) / raterange * upperRate
		d, err := DateToTime(dates[vid])
		if err != nil {
			logger.Printf("Error parsing date %s to time (video: %s)",
				dates[vid], vid)
			return nil, err
		}
		days := DaysBetween(d, now)
		s :=  freshDecay(viewscore + ratescore, days)
		tr := TrendingScore {
			VideoId: vid,
			Score: s,
			Date: d,
		}
		trendings = append(trendings, tr)
	}
	// sort the trendings in reverse order
	sort.Slice(trendings, func(i, j int) bool {
		if trendings[i].Score > trendings[j].Score {
			return true
		} else if trendings[i].Score < trendings[j].Score {
			return false
		} else {
			return trendings[i].Date.After(trendings[j].Date)
		}
	})
	return trendings, nil
}

// var r = rand.New(rand.NewSource(time.Now().UnixMilli()))
var r = rand.New(rand.NewSource(99))
func makeInfo() Info {
	allDates := make([]string, 0)
	allDates = append(allDates, "2022-03-01")
	allDates = append(allDates, "2022-03-02")
	allDates = append(allDates, "2022-03-03")
	allDates = append(allDates, "2022-03-04")
	allDates = append(allDates, "2022-03-05")
	allDates = append(allDates, "2022-03-06")
	allDates = append(allDates, "2022-03-07")
	allDates = append(allDates, "2022-03-08")

	s := r.Float64() * 5.0
	ok := false
	var sq float64 = 0
	for ; !ok; {
		sq = r.Float64() * 25.0
		if sq > s*s {
			ok = true
		}
	}
    i := Info {
    	VideoMeta: Meta {
        	UserId: "Justice",
        	Resolution: "1080p",
        	Duration: 30.5,
        	Description: "Fakes get out of academia!",
        	Date: allDates[r.Intn(len(allDates))],
    	},
    	// number of views
    	Views: r.Int63n(2000),
    	Rate: Rating {
			Num: r.Int63n(500),
			Score: s,
			ScoreSq: sq,
		},
    }
	return i
}

func makeNonRateInfo() Info {
	allDates := make([]string, 0)
	allDates = append(allDates, "2022-03-01")
	allDates = append(allDates, "2022-03-02")
	allDates = append(allDates, "2022-03-03")
	allDates = append(allDates, "2022-03-04")
	allDates = append(allDates, "2022-03-05")
	allDates = append(allDates, "2022-03-06")
	allDates = append(allDates, "2022-03-07")
	allDates = append(allDates, "2022-03-08")
    i := Info {
    	VideoMeta: Meta {
        	UserId: "Justice",
        	Resolution: "1080p",
        	Duration: 30.5,
        	Description: "Fakes get out of academia!",
        	Date: allDates[r.Intn(len(allDates))],
    	},
    	// number of views
    	Views: 0,
    	Rate: Rating {
			Num: 0,
			Score: 0.0,
			ScoreSq: 0.0,
		},
    }
	return i
}

func makeSingleRateInfo() Info {
	allDates := make([]string, 0)
	allDates = append(allDates, "2022-03-01")
	allDates = append(allDates, "2022-03-02")
	allDates = append(allDates, "2022-03-03")
	allDates = append(allDates, "2022-03-04")
	allDates = append(allDates, "2022-03-05")
	allDates = append(allDates, "2022-03-06")
	allDates = append(allDates, "2022-03-07")
	allDates = append(allDates, "2022-03-08")

	s := r.Float64() * 5.0
    i := Info {
    	VideoMeta: Meta {
        	UserId: "Justice",
        	Resolution: "1080p",
        	Duration: 30.5,
        	Description: "Fakes get out of academia!",
        	Date: allDates[r.Intn(len(allDates))],
    	},
    	// number of views
    	Views: 1,
    	Rate: Rating {
			Num: 1,
			Score: s,
			ScoreSq: s*s,
		},
    }
	return i
}

func showInfo(trendings []TrendingScore) {
	for _, tr := range(trendings) {
		logger.Printf("------------")
		logger.Printf("video: %s", tr.VideoId)
		logger.Printf("score: %.4f", tr.Score)
		logger.Printf("date: %s", TimeToDate(tr.Date))
	}
}

func main() {
	vinfo := make(map[string]Info)
    for i := 0; i < 20; i++ {
		vid := fmt.Sprintf("video-%d", i)
		vinfo[vid] = makeInfo()
	}
	tr, err := sortVideos(vinfo)
	fmt.Println(err)
	showInfo(tr)
	fmt.Println("#### --------------- ####")

	for i := 0; i < 20; i++ {
		toss := r.Intn(2)
		vid := fmt.Sprintf("corner-video-%d", i)
		if toss == 0 {
			vinfo[vid] = makeNonRateInfo()
		} else {
			vinfo[vid] = makeSingleRateInfo()
		}
	}
	tr, err = sortVideos(vinfo)
	fmt.Println(err)
	showInfo(tr)
}