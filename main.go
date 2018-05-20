package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

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
	path720 := string(ss[0] + ".720p.mp4")
	path576 := string(ss[0] + ".576p.mp4")
	path360 := string(ss[0] + ".360p.mp4")
	path240 := string(ss[0] + ".240p.mp4")
	path144 := string(ss[0] + ".144p.mp4")
	previewfile := string(previewpath + s[0] + ".png")
	SourceHeight := getResolution(path + file.Filename)
	//fmt.Println(SourceHeight)
	switch {
	case SourceHeight >= 1080:
		fmt.Println("Will code in 4 qualities")
	case SourceHeight >= 720:
		fmt.Println("Will code in 3 qualities")
	case SourceHeight >= 480:
		fmt.Println("Will code in 2 qualities")
	case SourceHeight >= 360:
		fmt.Println("Will code in 1 qualities")
	default:
		fmt.Print("less than, Will code in 1 qualities")

	}
	go transcode(path+file.Filename, path720, path576, path360, path240, path144, previewfile)
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
func transcode(path string, path720 string, path576 string, path360 string, path240 string, path144 string, preview string) {
	fmt.Println(path, preview)
	c720 := exec.Command("ffmpeg", "-i", path, "-c:v", "libx264", "-c:a", "aac", "-b:a", "128k", "-b:v", "2500k", "-preset:v", "veryfast", "-s", "1280x720", "-aspect", "16:9", "-f", "mp4", path720)
	c576 := exec.Command("ffmpeg", "-i", path, "-c:v", "libx264", "-c:a", "aac", "-b:a", "128k", "-b:v", "1500k", "-preset:v", "veryfast", "-s", "720x576", "-aspect", "16:9", "-f", "mp4", path576)
	c360 := exec.Command("ffmpeg", "-i", path, "-c:v", "libx264", "-c:a", "aac", "-b:a", "96k", "-b:v", "500k", "-preset:v", "veryfast", "-s", "480x360", "-aspect", "16:9", "-f", "mp4", path360)
	c240 := exec.Command("ffmpeg", "-i", path, "-c:v", "libx264", "-c:a", "aac", "-b:a", "96k", "-b:v", "250k", "-preset:v", "veryfast", "-s", "320x240", "-aspect", "16:9", "-f", "mp4", path240)
	c144 := exec.Command("ffmpeg", "-i", path, "-c:v", "libx264", "-c:a", "aac", "-b:a", "64k", "-b:v", "128k", "-preset:v", "veryfast", "-s", "192x144", "-aspect", "16:9", "-f", "mp4", path144)
	cprev := exec.Command("ffmpeg", "-i", path, "-an", "-ss", "00:00:00", "-an", "-r", "1", "-vframes", "1", "-s", "720x576", "-aspect", "16:9", preview)
	//crem := exec.Command("rm", "-f", path)
	err := c720.Start()
	if err != nil {
		fmt.Println("Error while transcoding yo 720p file: ", path, " ", err)
	}
	err = c576.Start()
	if err != nil {
		fmt.Println("Error while transcoding yo 720p file: ", path, " ", err)
	}
	err = c360.Start()
	if err != nil {
		fmt.Println("Error while transcoding yo 720p file: ", path, " ", err)
	}
	err = c240.Start()
	if err != nil {
		fmt.Println("Error while transcoding yo 720p file: ", path, " ", err)
	}
	err = c144.Start()
	if err != nil {
		fmt.Println("Error while transcoding yo 720p file: ", path, " ", err)
	}
	err = cprev.Start()
	if err != nil {
		fmt.Println("Error while generating preview file: ", path, " ", err)
	}
	/*err = crem.Start()
	if err != nil {
		fmt.Println("Error while removing: ", path, " ", err)
	}*/
}
func getResolution(file string) float64 {
	output, _ := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height", "-of", "json", file).Output()
	var jsonParsed *gabs.Container
	jsonParsed, _ = gabs.ParseJSON(output)
	ffprobeParsed, _ := jsonParsed.Path("streams").Children()
	children, _ := ffprobeParsed[0].ChildrenMap()
	height := children["height"].Data().(float64)
	return height
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
