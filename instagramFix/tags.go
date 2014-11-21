package instagramFix

import (
	"errors"
	"fmt"
	"github.com/carbocation/go-instagram/instagram"
	"net/url"
	"strconv"
)

type TagsService struct {
	Client *instagram.Client
}

// RecentMedia Get a list of recently tagged media.
//
// Instagram API docs: http://instagram.com/developer/endpoints/tags/#get_tags_media_recent
func (s *TagsService) RecentMediaFix(tagName string, opt *instagram.Parameters) ([]instagram.Media, *instagram.ResponsePagination, error) {
	u := fmt.Sprintf("tags/%v/media/recent", tagName)
	if opt != nil {
		params := url.Values{}
		if opt.Count != 0 {
			params.Add("count", strconv.FormatUint(opt.Count, 10))
		}
		if opt.MinID != "" {
			params.Add("min_id", opt.MinID)
		}
		if opt.MaxID != "" {
			params.Add("max_id", opt.MaxID)
		}
		u += "?" + params.Encode()
	}
	req, err := s.Client.NewRequest("GET", u, "")
	if err != nil {
		return nil, nil, err
	}

	media := new([]instagram.Media)

	_, err = s.Client.Do(req, media)
	if err != nil {
		if req != nil && req.URL != nil {
			return nil, nil, errors.New(fmt.Sprintf("go-instagram Tag.RecentMedia error:%s on URL %s", err.Error(), req.URL.String()))
		} else {
			return nil, nil, errors.New(fmt.Sprintf("go-instagram Tag.RecentMedia error:%s on nil URL", err.Error()))
		}
	}

	page := new(instagram.ResponsePagination)
	if s.Client.Response.Pagination != nil {
		page = s.Client.Response.Pagination
	}

	return *media, page, err
}
