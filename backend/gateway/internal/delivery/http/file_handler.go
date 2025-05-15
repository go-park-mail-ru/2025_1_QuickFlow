package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/sync/errgroup"

	"quickflow/gateway/internal/delivery/http/forms"
	errors2 "quickflow/gateway/internal/errors"
	http2 "quickflow/gateway/utils/http"
	"quickflow/shared/logger"
	"quickflow/shared/models"
)

type FileService interface {
	UploadManyFiles(ctx context.Context, files []*models.File) ([]string, error)
	DeleteFile(ctx context.Context, filename string) error
}

type FileHandler struct {
	fileService FileService
	policy      *bluemonday.Policy
}

func NewFileHandler(fileService FileService, policy *bluemonday.Policy) *FileHandler {
	return &FileHandler{
		fileService: fileService,
		policy:      policy,
	}
}

func (p *FileHandler) AddFiles(w http.ResponseWriter, r *http.Request) {
	// extracting user from context
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while adding files")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	logger.Info(ctx, fmt.Sprintf("User %s requested to add files", user.Username))

	// Parse the form data
	err := r.ParseMultipartForm(15 << 20) // 15 MB TODO
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to parse form: %s", err.Error()))
		http2.WriteJSONError(w, err)
		return
	}

	// Handle video files
	media, err := http2.GetFiles(r, "media")
	if errors.Is(err, http2.TooManyFilesErr) {
		logger.Error(ctx, fmt.Sprintf("Too many media files requested: %s", err.Error()))
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Too many media files requested", http.StatusBadRequest))
		return
	}
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to get video files: %s", err.Error()))
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get video files", http.StatusBadRequest))
		return
	}

	// Handle audio files
	audios, err := http2.GetFiles(r, "audio")
	if errors.Is(err, http2.TooManyFilesErr) {
		logger.Error(ctx, fmt.Sprintf("Too many audio files requested: %s", err.Error()))
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Too many audio files requested", http.StatusBadRequest))
		return
	}
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to get audio files: %s", err.Error()))
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get audio files", http.StatusBadRequest))
		return
	}

	// Handle other files
	otherFiles, err := http2.GetFiles(r, "files")
	if errors.Is(err, http2.TooManyFilesErr) {
		logger.Error(ctx, fmt.Sprintf("Too many files requested: %s", err.Error()))
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Too many files requested", http.StatusBadRequest))
		return
	}
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to get files: %s", err.Error()))
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get files", http.StatusBadRequest))
		return
	}

	var res forms.MessageAttachmentForm
	wg := errgroup.Group{}
	wg.Go(func() error {
		res.MediaURLs, err = p.fileService.UploadManyFiles(ctx, media)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Failed to upload media: %s", err.Error()))
			http2.WriteJSONError(w, err)
			return err
		}
		return nil
	})

	wg.Go(func() error {
		res.AudioURLs, err = p.fileService.UploadManyFiles(ctx, audios)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Failed to upload audio files: %s", err.Error()))
			http2.WriteJSONError(w, err)
			return err
		}
		return nil
	})

	wg.Go(func() error {
		res.FileURLs, err = p.fileService.UploadManyFiles(ctx, otherFiles)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Failed to upload other files: %s", err.Error()))
			http2.WriteJSONError(w, err)
			return err
		}
		return nil
	})

	if err := wg.Wait(); err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to upload files: %s", err.Error()))
		//http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to upload files", http.StatusInternalServerError))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(forms.PayloadWrapper[forms.MessageAttachmentForm]{Payload: res})
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to encode urls: %s", err.Error()))
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode urls", http.StatusInternalServerError))
		return
	}
}
