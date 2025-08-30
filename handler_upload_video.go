package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<30)

	defer r.Body.Close()

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

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "no video found in database", err)
		return
	}

	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "user unauthorized for video", err)
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}

	defer file.Close()

	ct := header.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(ct)
	if mediatype != "video/mp4" {
		respondWithError(w, http.StatusUnsupportedMediaType, "unsupported media type", err)
		return
	}

	tempFile, err := os.CreateTemp("", "tubely-mp4_upload")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to create temp file", err)
		return
	}

	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	_, err = io.Copy(tempFile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to copy to temp file", err)
		return
	}

	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to reset temp file pointer", err)
		return
	}

	aspectR, err := getVideoAspectRatio(tempFile.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to read temp file for aspect ratio", err)
		return
	}

	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to reset temp file pointer", err)
		return
	}

	FastStartTempFile, err := processVideoForFastStart(tempFile.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to process video for fast start", err)
		return
	}

	FSfile, err := os.Open(FastStartTempFile)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable open processed file", err)
		return
	}

	defer os.Remove(FSfile.Name())
	defer FSfile.Close()

	key := make([]byte, 32)
	rand.Read(key)
	randString := base64.RawURLEncoding.EncodeToString(key)
	VideoKey := fmt.Sprintf("%v/%v.mp4", aspectR, randString)

	PutObjectParams := s3.PutObjectInput{
		Bucket:      &cfg.s3Bucket,
		Key:         &VideoKey,
		Body:        FSfile,
		ContentType: &mediatype,
	}

	_, err = cfg.s3Client.PutObject(r.Context(), &PutObjectParams)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "failed to upload to AWS bucket", err)
		return
	}

	url := fmt.Sprintf("%v/%v", cfg.s3CfDistribution, VideoKey)
	video.VideoURL = &url
	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)

}
