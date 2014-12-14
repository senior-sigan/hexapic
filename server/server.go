package hexapic

import (
	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
	"encoding/json"
	"fmt"
	hexapic "github.com/blan4/hexapic/core"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	CLIENT_ID   string = "417c3ee8c9544530b83aa1c24de2abb3"
	UUID_LENGTH int    = 15
)

var (
	imageQuality    = jpeg.Options{Quality: jpeg.DefaultQuality}
	imageBadQuality = jpeg.Options{Quality: 50}
	index           []byte
	tagList         = [...]string{"cat", "dog", "lifeofadventure", "nya", "train"}
	letters         = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

type Tag struct {
	Value string
	Count int
}

type Question struct {
	UID      string
	Answer   string
	Variants []string
}

type QuestionResponse struct {
	UID      string   `json:"uid"`
	Variants []string `json:"variants"`
}

func init() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/stat", stat)
	http.HandleFunc("/location", location)
	http.HandleFunc("/random", random)
	index, _ = ioutil.ReadFile("index.html")
}

func generateUID() string {
	b := make([]rune, UUID_LENGTH)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func generateTags() (uid string, answer string, variants []string) {
	rand.Seed(time.Now().UTC().UnixNano())
	randList := rand.Perm(len(tagList))[0:4]
	log.Printf("Rand = %v", randList)
	answerIndex := rand.Int31n(4)
	answer = tagList[randList[answerIndex]]
	for _, index := range randList {
		variants = append(variants, tagList[index])
	}
	log.Printf("Tags = %v", variants)
	log.Printf("Answer = %v", answer)
	uid = generateUID()
	log.Printf("UID = %v", uid)

	return
}

func random(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	if key := r.FormValue("uid"); key != "" {
		k := datastore.NewKey(c, "Question", key, 0, nil)
		question := new(Question)

		httpClient := urlfetch.Client(c)
		api := hexapic.NewSearchApi(CLIENT_ID, httpClient)
		api.Count = 4

		if err := datastore.Get(c, k, question); err != nil {
			log.Printf("Can't find question for uid %v", key)
			http.Error(w, err.Error(), 404)
			return
		} else {
			if answer := r.FormValue("answer"); answer != "" {
				isRight := bool(answer == question.Answer)
				fmt.Fprintf(w, "{\"answer\":\"%s\",\"isRight\":\"%v\"}", answer, isRight)
				if isRight {
					datastore.Delete(c, k)
				}
				return
			}
			tag := question.Answer
			imgs := api.SearchByTag(tag)
			img := hexapic.GenerateCollage(imgs, 2, 2)
			w.Header().Set("Content-Type", "image/jpeg")
			jpeg.Encode(w, img, &imageBadQuality)
			return
		}
	}

	uid, answer, variants := generateTags()
	k := datastore.NewKey(c, "Question", uid, 0, nil)
	question := new(Question)

	datastore.Get(c, k, question)
	question.UID = uid
	question.Answer = answer
	question.Variants = variants

	if _, err := datastore.Put(c, k, question); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	resp := &QuestionResponse{
		UID:      question.UID,
		Variants: question.Variants,
	}
	data, _ := json.Marshal(resp)
	fmt.Fprintf(w, "%s", data)
}

func location(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	httpClient := urlfetch.Client(c)
	api := hexapic.NewSearchApi(CLIENT_ID, httpClient)
	var lat, lng float64
	located := 0

	if key, err := strconv.ParseFloat(r.FormValue("lat"), 64); err == nil {
		lat = key
		located++
	}
	if key, err := strconv.ParseFloat(r.FormValue("lng"), 64); err == nil {
		lng = key
		located++
	}

	if located != 2 {
		located = 0
		loc_data := r.Header.Get("X-Appengine-Citylatlong")
		if loc_data == "" {
			http.Error(w, "Can't find you", 500)
			return
		}

		loc := strings.Split(loc_data, ",")

		if key, err := strconv.ParseFloat(loc[0], 64); err == nil {
			lat = key
			located++
		}
		if key, err := strconv.ParseFloat(loc[1], 64); err == nil {
			lng = key
			located++
		}
		if located != 2 {
			http.Error(w, "Bad coordinates", 500)
			return
		}
	}

	height := 2
	width := 3
	imgs := api.SearchByLocation(lat, lng)
	img := hexapic.GenerateCollage(imgs, height, width)
	w.Header().Set("Content-Type", "image/jpeg")
	jpeg.Encode(w, img, &imageQuality)
}

func stat(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("Tag").Order("-Count")
	for t := q.Run(c); ; {
		var tag Tag
		_, err := t.Next(&tag)
		if err == datastore.Done {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprintf(w, "%v\n", tag)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	var imgs []image.Image
	height := 2
	width := 3
	c := appengine.NewContext(r)
	httpClient := urlfetch.Client(c)
	api := hexapic.NewSearchApi(CLIENT_ID, httpClient)

	if key, err := strconv.ParseInt(r.FormValue("height"), 10, 32); err == nil {
		if key < 6 && key > 0 {
			height = int(key)
		} else {
			http.Error(w, "Height to big. Must be from 0 to 5", 500)
			return
		}
	}

	if key, err := strconv.ParseInt(r.FormValue("width"), 10, 32); err == nil {
		if key < 6 && key > 0 {
			width = int(key)
		} else {
			http.Error(w, "Width to big. Must be from 0 to 5", 500)
			return
		}
	}

	api.Count = height * width

	if key := r.FormValue("tag"); key == "" {
		if key := r.FormValue("user"); key == "" {
			fmt.Fprintf(w, "%s", index)
			return
		} else {
			imgs = api.SearchByName(key)
		}
	} else {
		k := datastore.NewKey(c, "Tag", key, 0, nil)
		tag := new(Tag)

		if err := datastore.Get(c, k, tag); err != nil {
			tag.Value = key
			tag.Count = 1
		} else {
			tag.Count = tag.Count + 1
		}

		if _, err := datastore.Put(c, k, tag); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		imgs = api.SearchByTag(key)
	}

	img := hexapic.GenerateCollage(imgs, height, width)
	w.Header().Set("Content-Type", "image/jpeg")
	jpeg.Encode(w, img, &imageQuality)
}
