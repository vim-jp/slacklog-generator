package slacklog

import (
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/slack-go/slack"
)

// HostBySlack checks a file is hosted by slack or not.
func HostBySlack(f slack.File) bool {
	return strings.HasPrefix(f.URLPrivate, "https://files.slack.com/")
}

// LocalName returns name of local downloaded file.
func LocalName(f slack.File, url, suffix string) string {
	ext := filepath.Ext(url)
	nameExt := filepath.Ext(f.Name)
	name := f.Name[:len(f.Name)-len(ext)]
	if ext == "" {
		ext = nameExt
		if ext == "" {
			ext = FiletypeToExtension[f.Filetype]
		}
	}
	return RegulateFilename(name + suffix + ext)
}

// LocalPath returns path of local downloaded file.
func LocalPath(f slack.File) string {
	return path.Join(f.ID, url.PathEscape(LocalName(f, f.URLPrivate, "")))
}

// TopLevelMimetype extracts top level type from MIME Type.
func TopLevelMimetype(f slack.File) string {
	i := strings.Index(f.Mimetype, "/")
	if i < 0 {
		return ""
	}
	return f.Mimetype[:i]
}

// ThumbImagePath returns path of thumbnail image file.
func ThumbImagePath(f slack.File) string {
	if f.Thumb1024 == "" {
		return LocalPath(f)
	}
	return path.Join(f.ID, url.PathEscape(LocalName(f, f.Thumb1024, "_1024")))
}

// ThumbImageWidth returns width of thumbnail image.
func ThumbImageWidth(f slack.File) int {
	if f.Thumb1024 != "" {
		return f.Thumb1024W
	}
	return f.OriginalW
}

// ThumbImageHeight returns height of thumbnail image.
func ThumbImageHeight(f slack.File) int {
	if f.Thumb1024 != "" {
		return f.Thumb1024H
	}
	return f.OriginalH
}

// ThumbVideoPath returns local path of thumbnail for the video.
func ThumbVideoPath(f slack.File) string {
	return path.Join(f.ID, url.PathEscape(LocalName(f, f.ThumbVideo, "_thumb_video")))
}
