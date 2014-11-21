package hexapic

import (
	"appengine"
	"appengine/urlfetch"
	"fmt"
	hexapic "github.com/blan4/hexapic/core"
	"image/jpeg"
	"io/ioutil"
	"net/http"
)

var (
	imageQuality = jpeg.Options{Quality: jpeg.DefaultQuality}
	index        []byte
)

func init() {
	http.HandleFunc("/", handler)
	index, _ = ioutil.ReadFile("index.html")
}

func handler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	var (
		searchType string
		value      string
	)

	if key := r.FormValue("tag"); key == "" {
		if key := r.FormValue("user"); key == "" {
			fmt.Fprintf(w, "%s", index)
			return
		} else {
			searchType = "user"
			value = key
		}
	} else {
		searchType = "tag"
		value = key
	}

	client := urlfetch.Client(c)
	img, _ := hexapic.GetWallpaper(searchType, value, client)

	w.Header().Set("Content-Type", "image/jpeg")
	jpeg.Encode(w, img, &imageQuality)
}
