package web

import (
	"bufio"
	"encoding/binary"
	"fbi_installer/logging"
	"fmt"
	"io"
	"net"
	url2 "net/url"
	"os"
	"regexp"
	"time"

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

	return r.Run(listen)
}

func UploadCIA(c *gin.Context) {
	var filename = c.Query("filename")
	if !namePattern.MatchString(filename) {
		Message(c, 400, "filename must be ends with .cia, .3dsx, .cetk, .tik")
		return
	}

	targetPath := cleanFilename(filename)
	if targetPath == "" {
		Message(c, 400, "wrong filepath")
		return
	}

	writer, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
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
	dirs, err := os.ReadDir(DataDir)
	if err != nil {
		logger.Errorf("failed to list directory %s", DataDir)
		Message(c, 500)
		return
	}

	var fileNames []string
	for _, entry := range dirs {
		if !entry.IsDir() {
			fileNames = append(fileNames, entry.Name())
		}
	}

	c.JSON(200, fileNames)
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
		conn, err := net.Dial("tcp", form.Address+":5000")
		if err != nil {
			logger.Errorf("failed to open TCP connection to %s:5000, %v", form.Address, err)
			return
		}
		defer conn.Close()

		var url = fmt.Sprintf("http://%s/api/download?name=%s", hostname, url2.QueryEscape(form.Name))
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
