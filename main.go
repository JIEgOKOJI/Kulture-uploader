package main

import (
	//	"bufio"
	"bytes"
	"fmt"
	"io"
	//	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func getVideoProgress(c echo.Context) error {
	//todo VideoEncoding progress
	return nil
}
func upload(c echo.Context) error {

	//-----------
	// Read file
	//-----------

	// Source
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Destination
	s := strings.Split(file.Filename, ".")
	path := "/zfspool/video/" + s[0] + "/"
	previewpath := "/zfspool/previews/" + s[0] + "/"
	/*ext :=*/ exists(path)
	exists(previewpath)
	//if !ext {
	dst, err := os.Create(path + file.Filename)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}
	ss := strings.Split(path+file.Filename, ".")
	d := strings.Split(file.Filename, ".")
	path1080 := string(ss[0] + ".1080p.mp4")
	path720 := string(ss[0] + ".720p.mp4")
	path480 := string(ss[0] + ".480p.mp4")
	path360 := string(ss[0] + ".360p.mp4")
	previewfile := string(previewpath + s[0] + ".png")
	SourceHeight, SourceFrames := getResolution(path + file.Filename)
	fmt.Println(SourceFrames)
	var qualities int
	switch {
	case SourceHeight >= 1080:
		qualities = 4
		fmt.Println("Will code in 4 qualities")
	case SourceHeight >= 720:
		qualities = 3
		fmt.Println("Will code in 3 qualities")
	case SourceHeight >= 480:
		qualities = 2
		fmt.Println("Will code in 2 qualities")
	case SourceHeight >= 360:
		qualities = 1
		fmt.Println("Will code in 1 qualities")
	default:
		qualities = 1
		fmt.Print("less than, Will code in 1 qualities")

	}
	go transcode(path+file.Filename, qualities, path1080, path720, path480, path360, previewfile)
	manifest := "/video/" + d[0] + "/" + d[0] + ".,144p,240p,360p,576p,720p,.mp4.urlset/master.m3u8"
	playlists := make(map[string]string)
	playlists["720p"] = "/video/" + d[0] + "/" + d[0] + ".720p.mp4/index.m3u8"
	playlists["576p"] = "/video/" + d[0] + "/" + d[0] + ".576p.mp4/index.m3u8"
	playlists["360p"] = "/video/" + d[0] + "/" + d[0] + ".360p.mp4/index.m3u8"
	playlists["240p"] = "/video/" + d[0] + "/" + d[0] + ".240p.mp4/index.m3u8"
	playlists["144p"] = "/video/" + d[0] + "/" + d[0] + ".144p.mp4/index.m3u8"
	preview := "/previews/" + d[0] + "/" + d[0] + ".png"
	jsonObj := gabs.New()
	jsonObj.Set(true, "status")
	jsonObj.Set(file.Filename, "name")
	jsonObj.Set(manifest, "manifest")
	jsonObj.Set(playlists, "playlists")
	jsonObj.Set(preview, "preview")
	return c.HTML(http.StatusOK, fmt.Sprintf(jsonObj.String()))
	/*} else {
		jsonObj := gabs.New()
		jsonObj.Set(false, "status")
		return c.HTML(http.StatusOK, fmt.Sprintf(jsonObj.String()))
	}*/

}
func transcode(path string, qualities int, path1080 string, path720 string, path480 string, path360 string, preview string) {
	fmt.Println(path, preview)
	var stdoutBuf /*stderrBuf*/ bytes.Buffer
	var errStdout, errStderr error
	c1080 := exec.Command("ffmpeg", "-y", "-i", path, "-c:v", "libx264", "-c:a", "aac", "-b:a", "384k", "-b:v", "8000k", "-preset:v", "veryfast", "-s", "1920x1080", "-aspect", "16:9", "-f", "mp4", path1080)
	stdoutIn, _ := c1080.StdoutPipe()
	stderrIn, _ := c1080.StderrPipe()
	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	//	stderr := io.MultiWriter(os.Stderr, &stderrBuf)
	c1080.Start()
	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
	}()

	go func() {
		//_, errStderr = io.Copy(stderr, stderrIn)
		f, _ := os.Create("log")
		//w := //bufio.NewWriter(f)
		w := io.Writer(f)
		io.Copy(w, stderrIn)

	}()

	err := c1080.Wait()
	if err != nil {
		//fmt.Fatalf("cmd.Run() failed with %s\n", err)
	}
	if errStdout != nil || errStderr != nil {
		//fmt.Fatal("failed to capture stdout or stderr\n")
	}
	//outStr, _ := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	/*f, err := os.Create("log")
	w := bufio.NewWriter(f)
	fmt.Fprintf(w, stderrBuf.String())
	w.Flush()*/
	//c720, _ := exec.Command("ffmpeg", "-i", path, "-c:v", "libx264", "-c:a", "aac", "-b:a", "384k", "-b:v", "5000k", "-preset:v", "veryfast", "-s", "1280x720", "-aspect", "16:9", "-f", "mp4", path720).Output()
	//c480, _ := exec.Command("ffmpeg", "-i", path, "-c:v", "libx264", "-c:a", "aac", "-b:a", "384k", "-b:v", "2500k", "-preset:v", "veryfast", "-s", "720x480", "-aspect", "16:9", "-f", "mp4", path480).Output()
	//c360, _ := exec.Command("ffmpeg", "-i", path, "-c:v", "libx264", "-c:a", "aac", "-b:a", "384k", "-b:v", "1000k", "-preset:v", "veryfast", "-s", "480x360", "-aspect", "16:9", "-f", "mp4", path360).Output()
	//cprev := exec.Command("ffmpeg", "-i", path, "-an", "-ss", "00:00:00", "-an", "-r", "1", "-vframes", "1", "-s", "720x480", "-aspect", "16:9", preview)
	//fmt.Println(c1080)
}
func getResolution(file string) (float64, int) {
	output, _ := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height", "-of", "json", file).Output()
	var jsonParsed *gabs.Container
	jsonParsed, _ = gabs.ParseJSON(output)
	ffprobeParsed, _ := jsonParsed.Path("streams").Children()
	children, _ := ffprobeParsed[0].ChildrenMap()
	height := children["height"].Data().(float64)
	durationCmd, err := exec.Command("bash", "durationBash", file).Output()
	if err != nil {
		fmt.Println(err)
	}
	duration, _ := strconv.Atoi((string(durationCmd)))
	fmt.Println(string(durationCmd), err)
	fmt.Println(string(output))
	return height, duration
}
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		fmt.Println("Directory Exist")
		return true
	}
	if os.IsNotExist(err) {
		fmt.Println("Making Dir: " + path)
		merr := os.MkdirAll(path, os.ModePerm)
		if merr != nil {
			fmt.Println("Error making Dir: ", merr)
			return false
		}
		return false
	}
	return true
}
func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/", "public")
	e.POST("/upload", upload)

	e.Logger.Fatal(e.Start(":1323"))
}
