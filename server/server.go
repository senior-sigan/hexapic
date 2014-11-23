package hexapic

import (
	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
	"fmt"
	hexapic "github.com/blan4/hexapic/core"
	"image"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

const CLIENT_ID string = "417c3ee8c9544530b83aa1c24de2abb3"

var (
	imageQuality = jpeg.Options{Quality: jpeg.DefaultQuality}
	index        []byte
)

type Tag struct {
	Value string
	Count int
}

func init() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/stat", stat)
	http.HandleFunc("/location", location)
	index, _ = ioutil.ReadFile("index.html")
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
