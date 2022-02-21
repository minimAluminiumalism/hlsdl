package hlsdl

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/grafov/m3u8"
)

func parseHlsSegments(hlsURL string, headers map[string]string) ([]*Segment, error) {
	baseURL, err := url.Parse(hlsURL)
	if err != nil {
		return nil, errors.New("Invalid m3u8 url")
	}

	p, t, err := getM3u8ListType(hlsURL, headers)
	if err != nil {
		return nil, err
	}

	// Attention: this package doesn't support multiple stream playlist m3u8 file, raising error.
	if t != m3u8.MEDIA {
		var maxBandWidth uint32 = 0
		var bandWidth uint32
		var playlistURI string
		for _, playlistSeg := range p.(*m3u8.MasterPlaylist).Variants {
			bandWidth = playlistSeg.Bandwidth
			if bandWidth > maxBandWidth {
				maxBandWidth = bandWidth
				playlistURI = playlistSeg.URI
			}
		}

		if !strings.Contains(playlistURI, "http") {
			segmentURL, err := baseURL.Parse(playlistURI)
			if err != nil {
				return nil, err
			}
			playlistURI = segmentURL.String()
		}
		fmt.Println(playlistURI)
		segs, err := parseHlsSegments(playlistURI, headers)
		return segs, err
		// return nil, errors.New("No support the m3u8 format")
	}

	mediaList := p.(*m3u8.MediaPlaylist)
	segments := []*Segment{}
	for _, seg := range mediaList.Segments {
		if seg == nil {
			continue
		}

		if !strings.Contains(seg.URI, "http") {
			segmentURL, err := baseURL.Parse(seg.URI)
			if err != nil {
				return nil, err
			}

			seg.URI = segmentURL.String()
		}

		if seg.Key == nil && mediaList.Key != nil {
			seg.Key = mediaList.Key
		}

		if seg.Key != nil && !strings.Contains(seg.Key.URI, "http") {
			keyURL, err := baseURL.Parse(seg.Key.URI)
			if err != nil {
				return nil, err
			}

			seg.Key.URI = keyURL.String()
		}

		segment := &Segment{MediaSegment: seg}
		segments = append(segments, segment)
	}

	return segments, nil
}

func newRequest(url string, headers map[string]string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return req, nil
}

// m3u8.PlayList: m3u8 file content; m3u8.ListType: value `1` for master m3u8 file and `2` for normal media m3u8 file
// Type "master": https://multiplatform-f.akamaihd.net/i/multi/will/bunny/big_buck_bunny_,640x360_400,640x360_700,640x360_1000,950x540_1500,.f4v.csmil/master.m3u8
// Type "media": http://devimages.apple.com.edgekey.net/iphone/samples/bipbop/gear4/prog_index.m3u8

func getM3u8ListType(url string, headers map[string]string) (m3u8.Playlist, m3u8.ListType, error) {

	req, err := newRequest(url, headers)
	if err != nil {
		return nil, 0, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, 0, errors.New(res.Status)
	}

	p, t, err := m3u8.DecodeFrom(res.Body, false)
	if err != nil {
		return nil, 0, err
	}
	return p, t, nil
}
