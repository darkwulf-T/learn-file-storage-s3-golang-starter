package main

import (
	"fmt"
	"io"
	"net/http"

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

	// TODO: implement the upload here
	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
	}
	defer file.Close()
	contentType := header.Header.Get("Content-Type")

	data, err := io.ReadAll(file)

	metadata, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error getting the video", err)
		return
	}
	if userID != metadata.UserID {
		respondWithError(w, http.StatusUnauthorized, "You are not the owner of this Video", nil)
		return
	}
	newthumbnail := thumbnail{
		data:      data,
		mediaType: contentType,
	}
	videoThumbnails[videoID] = newthumbnail
	thumbnailURL := fmt.Sprintf("http://localhost:%s/api/thumbnails/%v", cfg.port, videoID)
	metadata.ThumbnailURL = &thumbnailURL
	err = cfg.db.UpdateVideo(metadata)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error updating video data", err)
		return
	}

	respondWithJSON(w, http.StatusOK, metadata)
}
