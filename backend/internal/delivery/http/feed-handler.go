package http

import (
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"quickflow/config"
	"quickflow/internal/delivery/forms"
	"quickflow/internal/models"
	http2 "quickflow/utils/http"
)

type PostUseCase interface {
	FetchFeed(ctx context.Context, user models.User, numPosts int, timestamp time.Time) ([]models.Post, error)
	AddPost(ctx context.Context, post models.Post) error
}

type FeedHandler struct {
	postUseCase PostUseCase
	authUseCase AuthUseCase
}

// NewFeedHandler creates new feed handler.
func NewFeedHandler(postUseCase PostUseCase, authUseCase AuthUseCase) *FeedHandler {
	return &FeedHandler{
		postUseCase: postUseCase,
		authUseCase: authUseCase,
	}
}

// AddPost добавляет новый пост
// @Summary Добавить пост
// @Description Добавляет новый пост в ленту
// @Tags Feed
// @Accept multipart/form-data
// @Produce json
// @Param text formData string true "Текст поста"
// @Param pics formData file false "Изображения"
// @Success 200 {string} string "OK"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 500 {object} forms.ErrorForm "Ошибка сервера"
// @Router /api/post [post]
func (f *FeedHandler) AddPost(w http.ResponseWriter, r *http.Request) {
	// extracting user from context
	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http2.WriteJSONError(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// parsing JSON
	var postForm forms.PostForm
	postForm.Text = r.FormValue("text")

	for _, fileHeaders := range r.MultipartForm.File["pics"] {
		var mimeType string
		if mimeType, err = detectMimeType(fileHeaders); err != nil {
			http2.WriteJSONError(w, "Failed to detect MIME type", http.StatusBadRequest)
			return
		}

		file, err := fileHeaders.Open()
		if err != nil {
			http2.WriteJSONError(w, "Failed to open file", http.StatusBadRequest)
			return
		}

		postForm.Images = append(postForm.Images, models.File{
			Reader:   file,
			Name:     fileHeaders.Filename,
			Size:     fileHeaders.Size,
			Ext:      filepath.Ext(fileHeaders.Filename),
			MimeType: mimeType,
		})

		file.Close()
	}

	post := postForm.ToPostModel(user.Id)

	err = f.postUseCase.AddPost(r.Context(), post)
	if err != nil {
		http2.WriteJSONError(w, "Failed to add post", http.StatusInternalServerError)
		return
	}
}

// GetFeed возвращает ленту постов
// @Summary Получить ленту
// @Description Возвращает список постов, опубликованных до указанного времени
// @Tags Feed
// @Produce json
// @Param posts_count query int true "Количество постов"
// @Param ts query string false "Временная метка"
// @Success 200 {array} forms.PostOut "Список постов"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 500 {object} forms.ErrorForm "Ошибка сервера"
// @Router /api/feed [get]
func (f *FeedHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	// extracting user from context
	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

	// parsing JSON
	var feedForm forms.FeedForm
	err := feedForm.GetParams(r.URL.Query())
	if err != nil {
		http2.WriteJSONError(w, "Failed to parse query params", http.StatusBadRequest)
		return
	}

	ts, err := time.Parse(config.TimeStampLayout, feedForm.Ts)
	if err != nil {
		ts = time.Now()
	}

	posts, err := f.postUseCase.FetchFeed(r.Context(), user, feedForm.Posts, ts)
	if err != nil {
		http2.WriteJSONError(w, "Failed to load feed", http.StatusInternalServerError)
		return
	}

	var postsOut []forms.PostOut
	for _, post := range posts {
		var postOut forms.PostOut
		postOut.FromPost(post)
		postsOut = append(postsOut, postOut)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(postsOut)
	if err != nil {
		http2.WriteJSONError(w, "Failed to encode feed", http.StatusInternalServerError)
		return
	}
}

// detectMimeType определяет MIME-тип файла, сначала проверяя заголовки, затем анализируя содержимое.
func detectMimeType(fileHeader *multipart.FileHeader) (string, error) {
	// Попробуем получить MIME-тип из заголовков
	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType != "" {
		return mimeType, nil
	}

	// Если в заголовках нет, пробуем определить по содержимому
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Читаем первые 512 байтов (это стандартный размер для определения типа)
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}

	return http.DetectContentType(buf[:n]), nil
}
