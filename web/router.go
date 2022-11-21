package web

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	url2 "net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"fbi_installer/html"
	"fbi_installer/logging"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
)

var DataDir = ""
var BaseURL = ""
var namePattern = regexp.MustCompile(`(?i)\.(cia|3dsx|cetk|tik)$`)

func StartGin(listen string) error {
	r := gin.New()
	r.Use(ginzap.Ginzap(logging.GetLogger(), time.RFC3339, false), gin.Recovery())
	r.GET("/ping", func(c *gin.Context) {
		c.Data(200, "text/plain", []byte("pong"))
	})
	r.POST("/api/upload", UploadCIA)
	r.GET("/api/list", ListCIA)
	r.POST("/api/send", SendTo3DS)
	r.GET("/api/download", DownloadCIA)
	r.DELETE("/api/delete", DeleteCIA)

	staticSub, _ := fs.Sub(html.StaticFS, "assets")
	r.StaticFS("/assets", http.FS(staticSub))
	r.StaticFileFS("/favicon.ico", "favicon.ico", http.FS(html.StaticFS))
	r.NoRoute(func(c *gin.Context) {
		c.FileFromFS("/", http.FS(html.StaticFS))
	})

	return r.Run(listen)
}

func UploadCIA(c *gin.Context) {
	defer c.Request.Body.Close()
	var filename = c.Query("name")
	if !namePattern.MatchString(filename) {
		Message(c, 400, "filename must be ends with .cia, .3dsx, .cetk, .tik")
		return
	}

	targetPath := cleanFilename(filename)
	if targetPath == "" {
		Message(c, 400, "wrong filepath")
		return
	}

	writer, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644) //nolint:gosec
	if err != nil {
		logger.Errorf("failed to open file %s", targetPath)
		Message(c, 500)
		return
	}
	defer writer.Close()

	reader := bufio.NewReader(c.Request.Body)
	_, err = io.Copy(writer, reader)
	if err != nil {
		logger.Errorf("failed to write file %s", targetPath)
		Message(c, 500)
		return
	}

	Message(c, 200, "success to upload %s", filename)
}

func ListCIA(c *gin.Context) {
	keyword := strings.ToLower(c.Query("s"))
	dirs, err := os.ReadDir(DataDir)
	if err != nil {
		logger.Errorf("failed to list directory %s", DataDir)
		Message(c, 500)
		return
	}

	var files []*GameFile
	for _, entry := range dirs {
		if !entry.IsDir() && strings.Contains(strings.ToLower(entry.Name()), keyword) {
			info, err := entry.Info()
			if err != nil {
				logger.Errorf("failed to get file %s info: %v", entry.Name(), err)
				continue
			}

			files = append(files, &GameFile{
				Name:    entry.Name(),
				Size:    info.Size(),
				ModTime: info.ModTime().Unix(),
			})
		}
	}

	c.JSON(200, files)
}

func DownloadCIA(c *gin.Context) {
	var name = c.Query("name")
	var targetPath = cleanFilename(name)
	if targetPath == "" {
		Message(c, 400, "filepath must be starts with %s", DataDir)
		return
	}

	c.File(targetPath)
}

func SendTo3DS(c *gin.Context) {
	var form Sender3DSForm
	if err := c.Bind(&form); err != nil {
		Message(c, 400)
		return
	}

	var hostname = c.Request.Host
	if BaseURL != "" {
		hostname = BaseURL
	}

	go func() {
		d := net.Dialer{Timeout: time.Second * 5}
		conn, err := d.Dial("tcp", form.Address+":5000")
		if err != nil {
			logger.Errorf("failed to open TCP connection to %s:5000, %v", form.Address, err)
			return
		}
		defer conn.Close()

		var url = fmt.Sprintf("%s/api/download?name=%s", hostname, url2.QueryEscape(form.Name))
		logger.Debugf("send %s to address %s", url, hostname)

		var bs = make([]byte, 4)
		binary.BigEndian.PutUint32(bs, uint32(len(url)))
		_, err = conn.Write(append(bs, []byte(url)...))
		if err != nil {
			logger.Errorf("failed to write data to %s, %v", form.Address, err)
			return
		}
	}()

	Message(c, 200)
}

func DeleteCIA(c *gin.Context) {
	var name = c.Query("name")
	var targetPath = cleanFilename(name)
	if !IsFile(targetPath) {
		Message(c, 404, "file %s not found", name)
		return
	}

	err := os.Remove(targetPath)
	if err != nil {
		Message(c, 500)
		logger.Errorf("failed to delete file %s: %v", targetPath, err)
		return
	}

	Message(c, 200)
}
