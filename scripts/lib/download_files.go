package slacklog

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

const downloadWorkerNum = 8

func doDownloadFiles() error {
	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		return fmt.Errorf("$SLACK_TOKEN required")
	}

	if len(os.Args) < 4 {
		fmt.Println("Usage: go run scripts/main.go download_files {log-dir} {files-dir}")
		return nil
	}

	logDir := filepath.Clean(os.Args[2])
	filesDir := filepath.Clean(os.Args[3])

	channels, _, err := readChannels(filepath.Join(logDir, "channels.json"), []string{"*"})
	if err != nil {
		return fmt.Errorf("could not read channels.json: %s", err)
	}

	if err := mkdir(filesDir); err != nil {
		return fmt.Errorf("could not create %s directory: %s", filesDir, err)
	}

	ch := make(chan *messageFile, downloadWorkerNum)
	wg := new(sync.WaitGroup)
	failed := false

	for i := 0; i < cap(ch); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for m := range ch {
				errs := m.downloadAll(filesDir, slackToken)
				for i := range errs {
					failed = true
					fmt.Fprintf(os.Stderr, "[error] Download failed: %s\n", errs[i])
				}
			}
		}()
	}

	for _, channel := range channels {
		messages, err := readAllMessages(filepath.Join(logDir, channel.Id))
		if err != nil {
			close(ch)
			return err
		}
		for _, message := range messages {
			for i := range message.Files {
				ch <- &message.Files[i]
			}
		}
	}

	close(ch)
	wg.Wait()

	if failed {
		return errors.New("failed to download some file(s)")
	}
	return nil
}

func urlToFilename(url string) string {
	i := strings.LastIndex(url, "/")
	if i < 0 {
		return ""
	}
	return url[i+1:]
}

func (f *messageFile) downloadURLsAndSuffixes() map[string]string {
	return map[string]string{
		f.UrlPrivate:   "",
		f.Thumb64:      "_64",
		f.Thumb80:      "_80",
		f.Thumb160:     "_160",
		f.Thumb360:     "_360",
		f.Thumb480:     "_480",
		f.Thumb720:     "_720",
		f.Thumb800:     "_800",
		f.Thumb960:     "_960",
		f.Thumb1024:    "_1024",
		f.Thumb360Gif:  "_360",
		f.Thumb480Gif:  "_480",
		f.DeanimateGif: "_deanimate_gif",
		f.ThumbVideo:   "_thumb_video",
	}
}

func (f *messageFile) downloadAll(outDir string, slackToken string) []error {
	fileBaseDir := path.Join(outDir, f.Id)
	err := mkdir(fileBaseDir)
	if err != nil {
		return []error{err}
	}

	var errs []error

	for url, suffix := range f.downloadURLsAndSuffixes() {
		err = f.downloadFile(fileBaseDir, url, suffix, slackToken)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (f *messageFile) downloadFilename(url, suffix string) string {
	ext := filepath.Ext(url)
	nameExt := filepath.Ext(f.Name)
	name := f.Name[:len(f.Name)-len(ext)]
	if ext == "" {
		ext = nameExt
		if ext == "" {
			ext = filetypeToExtension[f.Filetype]
		}
	}

	filename := strings.ReplaceAll(name+suffix+ext, "/", "_")

	// XXX: Jekyll doesn't publish files that name starts with some characters
	if strings.HasPrefix(filename, "_") || strings.HasPrefix(filename, ".") {
		filename = "files" + filename
	}

	return filename
}

func (f *messageFile) downloadFile(outDir, url, suffix, slackToken string) error {
	if url == "" {
		return nil
	}

	filename := f.downloadFilename(url, suffix)

	destFile := filepath.Join(outDir, filename)
	if _, err := os.Stat(destFile); err == nil {
		// Just skip already downloaded file
		return nil
	}

	fmt.Printf("Downloading: %s/%s [%s]\n", f.Id, filename, f.PrettyType)
	return downloadFile(url, destFile, slackToken)
}

func downloadFile(url, destFile, slackToken string) error {
	client := &http.Client{}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+slackToken)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("[%s]: %s", resp.Status, url)
	}

	w, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, resp.Body)

	return err
}

// https://api.slack.com/types/file
var filetypeToExtension = map[string]string{
	"auto":         "",             // Auto Detect Type,
	"text":         ".txt",         // Plain Text,
	"ai":           ".ai",          // Illustrator File,
	"apk":          ".apk",         // APK,
	"applescript":  ".applescript", // AppleScript,
	"binary":       "",             // Binary,
	"bmp":          ".bmp",         // Bitmap,
	"boxnote":      ".boxnote",     // BoxNote,
	"c":            ".c",           // C,
	"csharp":       ".cs",          // C#,
	"cpp":          ".cpp",         // C++,
	"css":          ".css",         // CSS,
	"csv":          ".csv",         // CSV,
	"clojure":      ".clj",         // Clojure,
	"coffeescript": ".coffee",      // CoffeeScript,
	"cfm":          ".cfm",         // ColdFusion,
	"d":            ".d",           // D,
	"dart":         ".dart",        // Dart,
	"diff":         ".diff",        // Diff,
	"doc":          ".doc",         // Word Document,
	"docx":         ".docx",        // Word document,
	"dockerfile":   ".dockerfile",  // Docker,
	"dotx":         ".dotx",        // Word template,
	"email":        ".eml",         // Email,
	"eps":          ".eps",         // EPS,
	"epub":         ".epub",        // EPUB,
	"erlang":       ".erl",         // Erlang,
	"fla":          ".fla",         // Flash FLA,
	"flv":          ".flv",         // Flash video,
	"fsharp":       ".fs",          // F#,
	"fortran":      ".f90",         // Fortran,
	"gdoc":         ".gdoc",        // GDocs Document,
	"gdraw":        ".gdraw",       // GDocs Drawing,
	"gif":          ".gif",         // GIF,
	"go":           ".go",          // Go,
	"gpres":        ".gpres",       // GDocs Presentation,
	"groovy":       ".groovy",      // Groovy,
	"gsheet":       ".gsheet",      // GDocs Spreadsheet,
	"gzip":         ".gz",          // GZip,
	"html":         ".html",        // HTML,
	"handlebars":   ".handlebars",  // Handlebars,
	"haskell":      ".hs",          // Haskell,
	"haxe":         ".hx",          // Haxe,
	"indd":         ".indd",        // InDesign Document,
	"java":         ".java",        // Java,
	"javascript":   ".js",          // JavaScript/JSON,
	"jpg":          ".jpeg",        // JPEG,
	"keynote":      ".keynote",     // Keynote Document,
	"kotlin":       ".kt",          // Kotlin,
	"latex":        ".tex",         // LaTeX/sTeX,
	"lisp":         ".lisp",        // Lisp,
	"lua":          ".lua",         // Lua,
	"m4a":          ".m4a",         // MPEG 4 audio,
	"markdown":     ".md",          // Markdown (raw),
	"matlab":       ".m",           // MATLAB,
	"mhtml":        ".mhtml",       // MHTML,
	"mkv":          ".mkv",         // Matroska video,
	"mov":          ".mov",         // QuickTime video,
	"mp3":          ".mp3",         // mp4,
	"mp4":          ".mp4",         // MPEG 4 video,
	"mpg":          ".mpeg",        // MPEG video,
	"mumps":        ".m",           // MUMPS,
	"numbers":      ".numbers",     // Numbers Document,
	"nzb":          ".nzb",         // NZB,
	"objc":         ".objc",        // Objective-C,
	"ocaml":        ".ml",          // OCaml,
	"odg":          ".odg",         // OpenDocument Drawing,
	"odi":          ".odi",         // OpenDocument Image,
	"odp":          ".odp",         // OpenDocument Presentation,
	"ods":          ".ods",         // OpenDocument Spreadsheet,
	"odt":          ".odt",         // OpenDocument Text,
	"ogg":          ".ogg",         // Ogg Vorbis,
	"ogv":          ".ogv",         // Ogg video,
	"pages":        ".pages",       // Pages Document,
	"pascal":       ".pp",          // Pascal,
	"pdf":          ".pdf",         // PDF,
	"perl":         ".pl",          // Perl,
	"php":          ".php",         // PHP,
	"pig":          ".pig",         // Pig,
	"png":          ".png",         // PNG,
	"post":         ".post",        // Slack Post,
	"powershell":   ".ps1",         // PowerShell,
	"ppt":          ".ppt",         // PowerPoint presentation,
	"pptx":         ".pptx",        // PowerPoint presentation,
	"psd":          ".psd",         // Photoshop Document,
	"puppet":       ".pp",          // Puppet,
	"python":       ".py",          // Python,
	"qtz":          ".qtz",         // Quartz Composer Composition,
	"r":            ".r",           // R,
	"rtf":          ".rtf",         // Rich Text File,
	"ruby":         ".rb",          // Ruby,
	"rust":         ".rs",          // Rust,
	"sql":          ".sql",         // SQL,
	"sass":         ".sass",        // Sass,
	"scala":        ".scala",       // Scala,
	"scheme":       ".scm",         // Scheme,
	"sketch":       ".sketch",      // Sketch File,
	"shell":        ".sh",          // Shell,
	"smalltalk":    ".st",          // Smalltalk,
	"svg":          ".svg",         // SVG,
	"swf":          ".swf",         // Flash SWF,
	"swift":        ".swift",       // Swift,
	"tar":          ".tar",         // Tarball,
	"tiff":         ".tiff",        // TIFF,
	"tsv":          ".tsv",         // TSV,
	"vb":           ".vb",          // VB.NET,
	"vbscript":     ".vbs",         // VBScript,
	"vcard":        ".vcf",         // vCard,
	"velocity":     ".vm",          // Velocity,
	"verilog":      ".v",           // Verilog,
	"wav":          ".wav",         // Waveform audio,
	"webm":         ".webm",        // WebM,
	"wmv":          ".wmv",         // Windows Media Video,
	"xls":          ".xls",         // Excel spreadsheet,
	"xlsx":         ".xlsx",        // Excel spreadsheet,
	"xlsb":         ".xlsb",        // Excel Spreadsheet (Binary, Macro Enabled),
	"xlsm":         ".xlsm",        // Excel Spreadsheet (Macro Enabled),
	"xltx":         ".xltx",        // Excel template,
	"xml":          ".xml",         // XML,
	"yaml":         ".yaml",        // YAML,
	"zip":          ".zip",         // Zip,
}
