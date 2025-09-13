package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
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

	// TODO: Develop here

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to read thumbnail from file", err)
		return
	}
	defer file.Close()

	// Get file type
	fileType := header.Header.Get("Content-Type")

	// Get video metadata and create object
	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to get video from db", err)
		return
	}

	// Generate unique file name
	b := make([]byte, 32) // 16 bytes = 128 bits
	_, err = rand.Read(b)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to encode fileName", err)
		return
	}
	encoded := base64.URLEncoding.EncodeToString(b)

	// Format the
	_, ext, _ := strings.Cut(fileType, "/") // "image", "png"
	fileName := fmt.Sprintf("assets/%s.%s", encoded, ext)

	writtenFile, err := os.Create(fileName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to write file to disk.", err)
		return
	}
	defer writtenFile.Close()

	// Copy contents
	io.Copy(writtenFile, file)

	// Adjust URL to hti the server
	adjustedFileName := fmt.Sprintf(
		"http://localhost:%s/assets/%s.%s",
		cfg.port,
		encoded,
		ext,
	)
	video.ThumbnailURL = &adjustedFileName

	// Update Video in the DB
	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
