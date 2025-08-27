package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const MaxMemory = 10 << 20
	r.ParseMultipartForm(MaxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}

	defer file.Close()

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "no video found in database", err)
		return
	}

	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "user unauthorized for video", err)
		return
	}

	ct := header.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(ct)
	if mediatype != "image/jpeg" && mediatype != "image/png" {
		respondWithError(w, http.StatusUnsupportedMediaType, "unsupported media type", err)
		return
	}
	parts := strings.Split(mediatype, "/")
	if len(parts) != 2 {
		respondWithError(w, http.StatusBadRequest, "invalid content-type", nil)
		return
	}
	ext := parts[1]

	var fileString string

	key := make([]byte, 32)
	rand.Read(key)
	randString := base64.RawURLEncoding.EncodeToString(key)

	if cfg.assetsRoot == "" {
		respondWithError(w, http.StatusBadRequest, "missing assetsRoot", nil)
		return
	} else {
		fileString = filepath.Join(cfg.assetsRoot, randString+"."+ext)
	}

	newfile, err := os.Create(fileString)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "file not created", err)
		return
	}

	defer newfile.Close()

	if _, err := io.Copy(newfile, file); err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to save to file", err)
		return
	}

	url := fmt.Sprintf("http://localhost:%v/assets/%v.%v", cfg.port, randString, ext)
	video.ThumbnailURL = &url
	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
