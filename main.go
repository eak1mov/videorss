package main

import (
	"context"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/SevereCloud/vksdk/v3/api"
	"github.com/SevereCloud/vksdk/v3/api/params"
	"github.com/SevereCloud/vksdk/v3/object"
	"golang.org/x/tools/blog/atom"
)

var (
	ErrorGroupNotAllowed  = errors.New("group not allowed")
	ErrorGroupNotFound    = errors.New("group not found")
	ErrorInvalidGroupName = errors.New("invalid group name")
	ErrorInvalidPassword  = errors.New("invalid password")
)

const maxGroupCount = 20 // length of whitelist from SettingsStorage
const bestImageHeight = 480

func createAtomEntry(group object.GroupsGroup, video object.VideoVideo) *atom.Entry {
	content := ""

	views := fmt.Sprint(video.Views)
	if video.Views > 1_000_000 {
		views = fmt.Sprintf("%.1fM", float64(video.Views)/1_000_000)
	} else if video.Views > 1_000 {
		views = fmt.Sprintf("%.0fK", float64(video.Views)/1_000)
	}

	content += "Duration: " + (time.Duration(video.Duration) * time.Second).String() + ", "
	content += "Views: " + views + ", "
	content += "Comments: " + fmt.Sprint(video.Comments) + ".<br>\n"
	content += "<br>\n"

	if len(video.Image) != 0 {
		heightDiff := func(img object.VideoVideoImage) int {
			d := int(img.Height) - bestImageHeight
			return max(d, -d)
		}
		image := slices.MinFunc(video.Image, func(a object.VideoVideoImage, b object.VideoVideoImage) int {
			return heightDiff(a) - heightDiff(b)
		})
		content += "<img src=\"" + image.URL + "\"><br>\n"
	}

	description := video.Description
	description = html.EscapeString(description)
	description = strings.Replace(description, "\n", "<br>", -1)
	content += description + "<br>\n"

	return &atom.Entry{
		Title:     video.Title,
		ID:        fmt.Sprint(video.ID),
		Link:      []atom.Link{{Rel: "alternate", Href: "https://vkvideo.ru/" + video.ToAttachment()}},
		Published: atom.Time(time.Unix(int64(video.Date), 0)),
		Updated:   atom.Time(time.Now()), // TODO: time when content was fetched from vk (from cache)
		Content:   &atom.Text{Type: "html", Body: content},
		Author:    &atom.Person{Name: group.ScreenName},
	}
}

func renderAtomFeed(group object.GroupsGroup, videos []object.VideoVideo) ([]byte, error) {
	feed := atom.Feed{
		Title:   group.Name,
		ID:      "vkvideo" + fmt.Sprint(group.ID),
		Updated: atom.Time(time.Now()), // TODO: time when content was fetched from vk (from cache)
		Link:    []atom.Link{{Rel: "alternate", Href: "https://vkvideo.ru/@" + group.ScreenName}},
	}
	for _, video := range videos {
		feed.Entry = append(feed.Entry, createAtomEntry(group, video))
	}
	data, err := xml.Marshal(&feed)
	if err != nil {
		return nil, err
	}
	return slices.Concat([]byte(xml.Header), data), nil
}

func processWall(group object.GroupsGroup, posts []object.WallWallpost) []object.VideoVideo {
	var videos []object.VideoVideo
	for _, post := range posts {
		if post.PostType != object.WallPostTypePost {
			continue // post, copy, reply, postpone, suggest
		}

		if post.CopyHistory != nil {
			continue // skip repost
		}

		if len(post.Attachments) != 1 {
			continue
		}

		attachment := post.Attachments[0]
		if attachment.Type != object.AttachmentTypeVideo {
			continue // video, audio, photo, note, etc
		}

		video := attachment.Video
		if video.Type != "video" {
			continue // video, music_video, movie, short_video, etc
		}

		// TODO: move custom checks to "filter" parameter
		if video.Title == "Video by "+group.Name {
			continue
		}
		if video.Title == "" {
			continue
		}

		videos = append(videos, video)
	}
	return videos
}

type wallResponse struct {
	Wall *api.WallGetExtendedResponse
	Err  error
}

func fetchWall(vk *api.VK, groupName string) *wallResponse {
	params := params.NewWallGetBuilder().Domain(groupName).Count(100).Params
	// we don't really care about request context,
	// we should update cache even if request was cancelled
	wall, err := vk.WallGetExtended(params.WithContext(context.Background()))

	if errors.Is(err, api.ErrParam) {
		return &wallResponse{nil, ErrorInvalidGroupName}
	}
	if err != nil {
		return &wallResponse{nil, err}
	}
	if len(wall.Groups) == 0 {
		return &wallResponse{nil, ErrorGroupNotFound}
	}

	return &wallResponse{&wall, nil}
}

type Server struct {
	vkApi        *api.VK
	wallCache    *Cache[*wallResponse]
	settings     SettingsStorage
	settingsAuth *SettingsAuth
}

func (s *Server) Vk(ctx context.Context, groupName string) ([]byte, error) {
	settings := s.settings.Get()
	// TODO: better settings format
	whitelist := strings.Split(settings, ";")

	if len(settings) == 0 || slices.Index(whitelist, groupName) == -1 {
		return nil, ErrorGroupNotAllowed
	}

	resp := s.wallCache.GetOrEmplace(groupName, func(string) *wallResponse {
		return fetchWall(s.vkApi, groupName)
	})
	if resp.Err != nil {
		return nil, resp.Err
	}

	videos := processWall(resp.Wall.Groups[0], resp.Wall.Items)
	data, err := renderAtomFeed(resp.Wall.Groups[0], videos)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Server) SettingsSalt() string {
	return s.settingsAuth.GenerateSalt()
}

func (s *Server) SettingsUpdate(data, hash, salt string) (string, error) {
	if !s.settingsAuth.CheckPassword(hash, salt) {
		return "", ErrorInvalidPassword
	}

	if len(data) == 0 {
		return s.settings.Get(), nil
	}

	err := s.settings.Put(data)
	if err != nil {
		return "", err
	}

	return data, nil
}

func loadSecret(env string, envFile string) string {
	value := os.Getenv(env)
	if value != "" {
		return value
	}

	filePath := os.Getenv(envFile)
	if filePath == "" {
		return ""
	}

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	return string(fileData)
}

func main() {
	vkMock := flag.Bool("mock", false, "use vk mock (for local tests)")
	flag.Parse()

	vkToken := loadSecret("VK_API_TOKEN", "VK_API_TOKEN_FILE")
	if vkToken == "" && !*vkMock {
		log.Fatal("VK_API_TOKEN not found")
	}

	vkTransport := http.DefaultTransport
	if *vkMock {
		vkTransport = NewVkMock()
	}
	vkTransport = NewThrottler(api.LimitUserToken, time.Second, vkTransport)

	vk := api.NewVK(vkToken)
	vk.Client = &http.Client{Transport: vkTransport}
	// do not use rate limiter from the library,
	// we already have Throttler with classic token bucket algorithm
	vk.Limit = 0

	cache := NewCache[*wallResponse](maxGroupCount, time.Hour)

	s3AccessKey := loadSecret("AWS_ACCESS_KEY_ID", "AWS_ACCESS_KEY_ID_FILE")
	s3Secret := loadSecret("AWS_SECRET_ACCESS_KEY", "AWS_SECRET_ACCESS_KEY_FILE")
	s3Bucket := os.Getenv("S3_BUCKET")
	s3ObjectKey := os.Getenv("S3_OBJECT_KEY")

	settingsFile := os.Getenv("SETTINGS_FILE")
	settingsPassword := loadSecret("SETTINGS_PASSWORD", "SETTINGS_PASSWORD_FILE")

	var settingsStorage SettingsStorage
	if s3AccessKey != "" && s3Secret != "" && s3Bucket != "" {
		settingsStorage = NewS3Storage(s3AccessKey, s3Secret, s3Bucket, s3ObjectKey)
	} else if settingsFile != "" {
		settingsStorage = NewFileStorage(settingsFile)
	} else {
		settingsStorage = NewMemoryStorage("")
	}

	server := &Server{
		vkApi:        vk,
		wallCache:    cache,
		settings:     settingsStorage,
		settingsAuth: NewSettingsAuth(settingsPassword),
	}

	startServerEcho(server)
}
