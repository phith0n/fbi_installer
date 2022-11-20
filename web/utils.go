package web

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"path/filepath"
	"strings"
)

func Message(c *gin.Context, code int, args ...interface{}) {
	if len(args) == 0 {
		c.AbortWithStatus(code)
	} else if len(args) == 1 {
		msg := args[0].(string)
		c.AbortWithStatusJSON(code, gin.H{
			"message": msg,
		})
	} else {
		msg := args[0].(string)
		c.AbortWithStatusJSON(code, gin.H{
			"message": fmt.Sprintf(msg, args[1:]...),
		})
	}
}

func cleanFilename(name string) string {
	targetPath, err := filepath.Abs(filepath.Join(DataDir, filepath.Clean(name)))
	if err != nil || !strings.HasPrefix(targetPath+string(os.PathSeparator), DataDir) {
		return ""
	}

	return targetPath
}

func IsFile(filename string) bool {
	if info, err := os.Stat(filename); err == nil {
		return !info.IsDir()
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		return false
	}
}
