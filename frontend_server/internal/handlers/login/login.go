package site_handler

import (
	"log/slog"
	"net/http"
	"fmt"
	"os"

	config "github.com/Leonid-Sarmatov/golang-yandex-last-fight/frontend_server/internal/config"
)

func NewLoginSiteHandler(logger *slog.Logger, config *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filePath := config.LoginPagePath

		file, err := os.Open(filePath)
		if err != nil {
			logger.Error("Failed to open file", err.Error())
			http.Error(w, "Failed to open file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			logger.Error("Failed to get file info", err.Error())
			http.Error(w, "Failed to get file info", http.StatusInternalServerError)
			return
		}

		fileSize := fileInfo.Size()
		buffer := make([]byte, fileSize)

		_, err = file.Read(buffer)
		if err != nil {
			logger.Error("Failed to read file", err.Error())
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			return
		}

		fmt.Fprint(w, string(buffer))
	}
}
