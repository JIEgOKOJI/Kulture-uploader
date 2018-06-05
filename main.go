package main

import (
	//	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func getVideoSize(c echo.Context) error {
	requestPath := strings.Split(c.Request().URL.String(), "/")
	id := requestPath[2]
	path := "/zfspool/video/" + id
	ext := exists(path)
	files, err := filepath.Glob(path + "/*.mp4")
	if err != nil {
		fmt.Println(err)
	}
	if ext {
		filesArray := make(map[string]int64)
		for _, value := range files {
			file, err := os.Open(value)
			if err != nil {
				// handle the error here
				return err
			}
			defer file.Close()

			// get the file size
			stat, err := file.Stat()
			if err != nil {
				return err
			}
			filename := strings.Split(value, "/")
			fmt.Println("File", filename[4], "size is ", stat.Size())
			filesArray[filename[4]] = stat.Size()

		}
		jsonObj := gabs.New()
		jsonObj.Set(filesArray, "files")
		return c.HTML(http.StatusOK, fmt.Sprintf(jsonObj.String()))
	}
	jsonObj := gabs.New()
	jsonObj.Set("notfound", "files")
	return c.HTML(http.StatusOK, fmt.Sprintf(jsonObj.String()))
}
func getVideoProgress(c echo.Context) error {
	requestPath := strings.Split(c.Request().URL.String(), "/")
	id := requestPath[2]
	path := "/zfspool/video/" + id + "/log"
	ext := exists(path)
	fmt.Println(path, ext)
	if ext {
		dat, err := ioutil.ReadFile(path + "/done.log")
		if err == nil {
			return c.HTML(http.StatusOK, fmt.Sprintf(string(dat)))
		}

	}
	return nil
}
func uploadPreview(c echo.Context) error {
	requestPath := strings.Split(c.Request().URL.String(), "/")
	id := requestPath[2]
	jsonObj := gabs.New()
	jsonObj.Set(id, "name")
	result := true
	jsonObj.Set(result, "result")
	file, err := c.FormFile("file")
	if err != nil {
		result = false
		jsonObj.Set(result, "result")
		return c.HTML(http.StatusOK, fmt.Sprintf(jsonObj.String()))
	}
	src, err := file.Open()
	if err != nil {
		result = false
		jsonObj.Set(result, "result")
		return c.HTML(http.StatusOK, fmt.Sprintf(jsonObj.String()))
	}
	defer src.Close()
	path := "/zfspool/previews/" + id + "/"
	dst, err := os.Create(path + "4.png")
	if err != nil {
		result = false
		jsonObj.Set(result, "result")
		return c.HTML(http.StatusOK, fmt.Sprintf(jsonObj.String()))
	}
	defer dst.Close()
	if _, err = io.Copy(dst, src); err != nil {
		result = false
		jsonObj.Set(result, "result")
		return c.HTML(http.StatusOK, fmt.Sprintf(jsonObj.String()))
	}
	return c.HTML(http.StatusOK, fmt.Sprintf(jsonObj.String()))
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
	idFileName := rand_str(8) + "." + s[1]
	idFileNameOnly := strings.Split(idFileName, ".")
	path := "/zfspool/video/" + idFileNameOnly[0] + "/"

	/*ext :=*/
	existsAndMake(path)

	//if !ext {

	dst, err := os.Create(path + idFileName)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}
	ss := strings.Split(path+idFileName, ".")
	d := strings.Split(idFileName, ".")
	previewpath := "/zfspool/previews/" + d[0] + "/"
	existsAndMake(previewpath)
	path1080 := string(ss[0] + ".1080p.mp4")
	path720 := string(ss[0] + ".720p.mp4")
	path480 := string(ss[0] + ".480p.mp4")
	path360 := string(ss[0] + ".360p.mp4")
	//	previewfile := string(previewpath + s[0] + ".png")
	SourceHeight, duration := getResolution(path + idFileName)
	var qualities int
	playlists := make(map[string]string)
	previews := make(map[int]string)
	var manifest string
	switch {
	case SourceHeight >= 1080:
		qualities = 4
		playlists["1080p"] = "/video/" + d[0] + "/" + d[0] + ".1080p.mp4/index.m3u8"
		playlists["720p"] = "/video/" + d[0] + "/" + d[0] + ".720p.mp4/index.m3u8"
		playlists["480p"] = "/video/" + d[0] + "/" + d[0] + ".480p.mp4/index.m3u8"
		playlists["360p"] = "/video/" + d[0] + "/" + d[0] + ".360p.mp4/index.m3u8"
		manifest = "/video/" + d[0] + "/" + d[0] + ".,360p,480p,720p,1080p,.mp4.urlset/master.m3u8"
		fmt.Println("Will code in 4 qualities")
	case SourceHeight >= 720:
		qualities = 3
		playlists["720p"] = "/video/" + d[0] + "/" + d[0] + ".720p.mp4/index.m3u8"
		playlists["480p"] = "/video/" + d[0] + "/" + d[0] + ".480p.mp4/index.m3u8"
		playlists["360p"] = "/video/" + d[0] + "/" + d[0] + ".360p.mp4/index.m3u8"
		manifest = "/video/" + d[0] + "/" + d[0] + ".,360p,480p,720p,.mp4.urlset/master.m3u8"
		fmt.Println("Will code in 3 qualities")
	case SourceHeight >= 480:
		playlists["480p"] = "/video/" + d[0] + "/" + d[0] + ".480p.mp4/index.m3u8"
		playlists["360p"] = "/video/" + d[0] + "/" + d[0] + ".360p.mp4/index.m3u8"
		manifest = "/video/" + d[0] + "/" + d[0] + ".,360p,480p,.mp4.urlset/master.m3u8"
		qualities = 2
		fmt.Println("Will code in 2 qualities")
	case SourceHeight >= 360:
		qualities = 1
		playlists["360p"] = "/video/" + d[0] + "/" + d[0] + ".360p.mp4/index.m3u8"
		manifest = "/video/" + d[0] + "/" + d[0] + ".,360p,.mp4.urlset/master.m3u8"
		fmt.Println("Will code in 1 qualities")
	default:
		qualities = 1
		playlists["360p"] = "/video/" + d[0] + "/" + d[0] + ".360p.mp4/index.m3u8"
		manifest = "/video/" + d[0] + "/" + d[0] + ".,360p,.mp4.urlset/master.m3u8"
		fmt.Print("less than, Will code in 1 qualities")

	}
	go transcode(path+idFileName, qualities, path1080, path720, path480, path360, previewpath, path+"log", idFileName, duration)
	i := 0
	for i < 5 {
		previews[i] = previewpath + strconv.Itoa(i) + ".png"
		i++
	}
	jsonObj := gabs.New()
	jsonObj.Set(true, "status")
	jsonObj.Set(d[0], "name")
	jsonObj.Set(manifest, "manifest")
	jsonObj.Set(playlists, "playlists")
	jsonObj.Set(previews, "preview")
	return c.HTML(http.StatusOK, fmt.Sprintf(jsonObj.String()))
	/*} else {
		jsonObj := gabs.New()
		jsonObj.Set(false, "status")
		return c.HTML(http.StatusOK, fmt.Sprintf(jsonObj.String()))
	}*/

}
func transcode(path string, qualities int, path1080 string, path720 string, path480 string, path360 string, preview string, logpath string, name string, duration int) {
	fmt.Println(path, preview)
	switch qualities {
	case 4:
		//transcode 1080
		func() {
			writePrgoress(4, 0, logpath, name)
			var stdoutBuf bytes.Buffer
			var errStdout, errStderr error
			c1080 := exec.Command("ffmpeg", "-y", "-i", path, "-c:v", "libx264", "-c:a", "aac", "-b:a", "384k", "-b:v", "8000k", "-preset:v", "veryfast", "-s", "1920x1080", "-aspect", "16:9", "-f", "mp4", path1080)
			stdoutIn, _ := c1080.StdoutPipe()
			stderrIn, _ := c1080.StderrPipe()
			stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
			c1080.Start()
			go func() {
				_, errStdout = io.Copy(stdout, stdoutIn)
			}()
			go func() {
				f, _ := os.Create(logpath + "/" + name + ".1080.log")
				w := io.Writer(f)
				io.Copy(w, stderrIn)

			}()
			err := c1080.Wait()
			if err != nil {
				fmt.Println("cmd.Run() failed with %s\n", err)
			}
			if errStdout != nil || errStderr != nil {
				fmt.Println("failed to capture stdout or stderr\n")
			}
			writePrgoress(4, 1, logpath, name)
		}()
		//transcode 720
		func() {
			var stdoutBuf bytes.Buffer
			var errStdout, errStderr error
			c720 := exec.Command("ffmpeg", "-y", "-i", path, "-c:v", "libx264", "-c:a", "aac", "-b:a", "384k", "-b:v", "5000k", "-preset:v", "veryfast", "-s", "1280x720", "-aspect", "16:9", "-f", "mp4", path720)
			stdoutIn, _ := c720.StdoutPipe()
			stderrIn, _ := c720.StderrPipe()
			stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
			c720.Start()
			go func() {
				_, errStdout = io.Copy(stdout, stdoutIn)
			}()
			go func() {
				f, _ := os.Create(logpath + "/" + name + ".720.log")
				w := io.Writer(f)
				io.Copy(w, stderrIn)

			}()
			err := c720.Wait()
			if err != nil {
				fmt.Println("cmd.Run() failed with %s\n", err)
			}
			if errStdout != nil || errStderr != nil {
				fmt.Println("failed to capture stdout or stderr\n")
			}
			writePrgoress(4, 2, logpath, name)
		}()
		//transcode 480
		func() {
			var stdoutBuf bytes.Buffer
			var errStdout, errStderr error
			c480 := exec.Command("ffmpeg", "-y", "-i", path, "-c:v", "libx264", "-c:a", "aac", "-b:a", "384k", "-b:v", "2500k", "-preset:v", "veryfast", "-s", "720x480", "-aspect", "16:9", "-f", "mp4", path480)
			stdoutIn, _ := c480.StdoutPipe()
			stderrIn, _ := c480.StderrPipe()
			stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
			c480.Start()
			go func() {
				_, errStdout = io.Copy(stdout, stdoutIn)
			}()
			go func() {
				f, _ := os.Create(logpath + "/" + name + ".480.log")
				w := io.Writer(f)
				io.Copy(w, stderrIn)

			}()
			err := c480.Wait()
			if err != nil {
				fmt.Println("cmd.Run() failed with %s\n", err)
			}
			if errStdout != nil || errStderr != nil {
				fmt.Println("failed to capture stdout or stderr\n")
			}
			writePrgoress(4, 3, logpath, name)
		}()
		//transcode 360
		func() {
			var stdoutBuf bytes.Buffer
			var errStdout, errStderr error
			c360 := exec.Command("ffmpeg", "-y", "-i", path, "-c:v", "libx264", "-c:a", "aac", "-b:a", "384k", "-b:v", "1000k", "-preset:v", "veryfast", "-s", "480x360", "-aspect", "16:9", "-f", "mp4", path360)
			stdoutIn, _ := c360.StdoutPipe()
			stderrIn, _ := c360.StderrPipe()
			stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
			c360.Start()
			go func() {
				_, errStdout = io.Copy(stdout, stdoutIn)
			}()
			go func() {
				f, _ := os.Create(logpath + "/" + name + ".360.log")
				w := io.Writer(f)
				io.Copy(w, stderrIn)

			}()
			err := c360.Wait()
			if err != nil {
				fmt.Println("cmd.Run() failed with %s\n", err)
			}
			if errStdout != nil || errStderr != nil {
				fmt.Println("failed to capture stdout or stderr\n")
			}
			writePrgoress(4, 4, logpath, name)
		}()
	case 3:
		//transcode 720
		func() {
			writePrgoress(3, 0, logpath, name)
			var stdoutBuf bytes.Buffer
			var errStdout, errStderr error
			c720 := exec.Command("ffmpeg", "-y", "-i", path, "-c:v", "libx264", "-c:a", "aac", "-b:a", "384k", "-b:v", "5000k", "-preset:v", "veryfast", "-s", "1280x720", "-aspect", "16:9", "-f", "mp4", path720)
			stdoutIn, _ := c720.StdoutPipe()
			stderrIn, _ := c720.StderrPipe()
			stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
			c720.Start()
			go func() {
				_, errStdout = io.Copy(stdout, stdoutIn)
			}()
			fmt.Println(logpath + "/" + name + ".720.log")
			go func() {
				f, _ := os.Create(logpath + "/" + name + ".720.log")

				w := io.Writer(f)
				io.Copy(w, stderrIn)

			}()
			err := c720.Wait()
			if err != nil {
				fmt.Println("cmd.Run() failed with %s\n", err)
			}
			if errStdout != nil || errStderr != nil {
				fmt.Println("failed to capture stdout or stderr\n")
			}
			writePrgoress(3, 1, logpath, name)
		}()
		//transcode 480
		func() {
			var stdoutBuf bytes.Buffer
			var errStdout, errStderr error
			c480 := exec.Command("ffmpeg", "-y", "-i", path, "-c:v", "libx264", "-c:a", "aac", "-b:a", "384k", "-b:v", "2500k", "-preset:v", "veryfast", "-s", "720x480", "-aspect", "16:9", "-f", "mp4", path480)
			stdoutIn, _ := c480.StdoutPipe()
			stderrIn, _ := c480.StderrPipe()
			stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
			c480.Start()
			go func() {
				_, errStdout = io.Copy(stdout, stdoutIn)
			}()
			go func() {
				f, _ := os.Create(logpath + "/" + name + ".480.log")
				w := io.Writer(f)
				io.Copy(w, stderrIn)

			}()
			err := c480.Wait()
			if err != nil {
				fmt.Println("cmd.Run() failed with %s\n", err)
			}
			if errStdout != nil || errStderr != nil {
				fmt.Println("failed to capture stdout or stderr\n")
			}
			writePrgoress(3, 2, logpath, name)
		}()
		//transcode 360
		func() {
			var stdoutBuf bytes.Buffer
			var errStdout, errStderr error
			c360 := exec.Command("ffmpeg", "-y", "-i", path, "-c:v", "libx264", "-c:a", "aac", "-b:a", "384k", "-b:v", "1000k", "-preset:v", "veryfast", "-s", "480x360", "-aspect", "16:9", "-f", "mp4", path360)
			stdoutIn, _ := c360.StdoutPipe()
			stderrIn, _ := c360.StderrPipe()
			stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
			c360.Start()
			go func() {
				_, errStdout = io.Copy(stdout, stdoutIn)
			}()
			go func() {
				f, _ := os.Create(logpath + "/" + name + ".360.log")
				w := io.Writer(f)
				io.Copy(w, stderrIn)

			}()
			err := c360.Wait()
			if err != nil {
				fmt.Println("cmd.Run() failed with %s\n", err)
			}
			if errStdout != nil || errStderr != nil {
				fmt.Println("failed to capture stdout or stderr\n")
			}
			writePrgoress(3, 3, logpath, name)
		}()
	case 2:
		//transcode 480
		func() {
			writePrgoress(2, 0, logpath, name)
			var stdoutBuf bytes.Buffer
			var errStdout, errStderr error
			c480 := exec.Command("ffmpeg", "-y", "-i", path, "-c:v", "libx264", "-c:a", "aac", "-b:a", "384k", "-b:v", "2500k", "-preset:v", "veryfast", "-s", "720x480", "-aspect", "16:9", "-f", "mp4", path480)
			stdoutIn, _ := c480.StdoutPipe()
			stderrIn, _ := c480.StderrPipe()
			stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
			c480.Start()
			go func() {
				_, errStdout = io.Copy(stdout, stdoutIn)
			}()
			go func() {
				f, _ := os.Create(logpath + "/" + name + ".480.log")
				w := io.Writer(f)
				io.Copy(w, stderrIn)

			}()
			err := c480.Wait()
			if err != nil {
				fmt.Println("cmd.Run() failed with %s\n", err)
			}
			if errStdout != nil || errStderr != nil {
				fmt.Println("failed to capture stdout or stderr\n")
			}
			writePrgoress(2, 1, logpath, name)
		}()
		//transcode 360
		func() {
			var stdoutBuf bytes.Buffer
			var errStdout, errStderr error
			c360 := exec.Command("ffmpeg", "-y", "-i", path, "-c:v", "libx264", "-c:a", "aac", "-b:a", "384k", "-b:v", "1000k", "-preset:v", "veryfast", "-s", "480x360", "-aspect", "16:9", "-f", "mp4", path360)
			stdoutIn, _ := c360.StdoutPipe()
			stderrIn, _ := c360.StderrPipe()
			stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
			c360.Start()
			go func() {
				_, errStdout = io.Copy(stdout, stdoutIn)
			}()
			go func() {
				f, _ := os.Create(logpath + "/" + name + ".360.log")
				w := io.Writer(f)
				io.Copy(w, stderrIn)

			}()
			err := c360.Wait()
			if err != nil {
				fmt.Println("cmd.Run() failed with %s\n", err)
			}
			if errStdout != nil || errStderr != nil {
				fmt.Println("failed to capture stdout or stderr\n")
			}
			writePrgoress(2, 2, logpath, name)
		}()
	case 1:
		//transcode 360
		func() {
			writePrgoress(1, 0, logpath, name)
			var stdoutBuf bytes.Buffer
			var errStdout, errStderr error
			c360 := exec.Command("ffmpeg", "-y", "-i", path, "-c:v", "libx264", "-c:a", "aac", "-b:a", "384k", "-b:v", "1000k", "-preset:v", "veryfast", "-s", "480x360", "-aspect", "16:9", "-f", "mp4", path360)
			stdoutIn, _ := c360.StdoutPipe()
			stderrIn, _ := c360.StderrPipe()
			stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
			c360.Start()
			go func() {
				_, errStdout = io.Copy(stdout, stdoutIn)
			}()
			go func() {
				f, _ := os.Create(logpath + "/" + name + ".360.log")
				w := io.Writer(f)
				io.Copy(w, stderrIn)

			}()
			err := c360.Wait()
			if err != nil {
				fmt.Println("cmd.Run() failed with %s\n", err)
			}
			if errStdout != nil || errStderr != nil {
				fmt.Println("failed to capture stdout or stderr\n")
			}
			writePrgoress(1, 1, logpath, name)
		}()
	}
	//genpreview
	func() {
		cprev, err := exec.Command("ffmpeg", "-y", "-i", path, "-an", "-ss", strconv.Itoa(duration-30), "-an", "-r", "1", "-vframes", "1", "-s", "960x540", "-aspect", "16:9", preview+"0.png").Output()
		fmt.Println(cprev, err)
		cprev, err = exec.Command("ffmpeg", "-y", "-i", path, "-an", "-ss", strconv.Itoa(duration/2), "-an", "-r", "1", "-vframes", "1", "-s", "960x540", "-aspect", "16:9", preview+"1.png").Output()
		fmt.Println(cprev, err)
		cprev, err = exec.Command("ffmpeg", "-y", "-i", path, "-an", "-ss", strconv.Itoa(duration/3), "-an", "-r", "1", "-vframes", "1", "-s", "960x540", "-aspect", "16:9", preview+"2.png").Output()
		fmt.Println(cprev, err)
		cprev, err = exec.Command("ffmpeg", "-y", "-i", path, "-an", "-ss", strconv.Itoa(duration/4), "-an", "-r", "1", "-vframes", "1", "-s", "960x540", "-aspect", "16:9", preview+"3.png").Output()
		fmt.Println(cprev, err)
		cprev, err = exec.Command("ffmpeg", "-y", "-i", path, "-an", "-ss", strconv.Itoa(30), "-an", "-r", "1", "-vframes", "1", "-s", "960x540", "-aspect", "16:9", preview+"4.png").Output()
		fmt.Println(cprev, err)
	}()
	//removeSource
	func() {
		err := os.Remove(path)
		fmt.Println(err)
	}()
}
func writePrgoress(need int, done int, path string, name string) {
	percentage := 100 / need
	percentage = percentage * done
	fmt.Println(percentage)
	jsonObj := gabs.New()
	jsonObj.Set("inprogress", "status")
	if percentage == 99 {
		percentage = 100
		jsonObj.Set("success", "status")
	}
	if percentage == 100 {
		jsonObj.Set("success", "status")
	}
	jsonObj.Set(name, "filename")
	jsonObj.Set(percentage, "done")

	fmt.Println(jsonObj.String())
	err := ioutil.WriteFile(path+"/done.log", jsonObj.Bytes(), 0644)
	if err != nil {
		fmt.Println(err)
	}
}
func getResolution(file string) (float64, int) {
	output, _ := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height,duration", "-of", "json", file).Output()
	var jsonParsed *gabs.Container
	jsonParsed, _ = gabs.ParseJSON(output)
	ffprobeParsed, _ := jsonParsed.Path("streams").Children()
	children, _ := ffprobeParsed[0].ChildrenMap()
	height := children["height"].Data().(float64)
	duration := children["duration"].Data().(string)
	d := strings.Split(duration, ".")
	durationInt, err := strconv.Atoi(d[0])
	fmt.Println(durationInt, err)
	return height, durationInt
}
func existsAndMake(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		fmt.Println("Directory Exist")
		return true
	}
	if os.IsNotExist(err) {
		fmt.Println("Making Dir: " + path)
		os.MkdirAll(path+"/log", os.ModePerm)
		merr := os.MkdirAll(path, os.ModePerm)
		if merr != nil {
			fmt.Println("Error making Dir: ", merr)
			return false
		}
		return false
	}
	return true
}
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		fmt.Println("Directory Exist")
		return true
	}

	return false

}
func rand_str(str_size int) string {
	alphanum := "0123456789abcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, str_size)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}
func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/", "public")
	e.POST("/upload", upload)
	e.POST("/uploadPreview/*", uploadPreview)
	e.GET("/getVideoProgress/*", getVideoProgress)
	e.GET("/getVideoSize/*", getVideoSize)
	//e.GET("/uploadPreview", uploadPreview)

	e.Logger.Fatal(e.Start(":1323"))
}
