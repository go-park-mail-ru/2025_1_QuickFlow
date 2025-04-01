package http

import (
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"quickflow/internal/delivery/forms"
	"quickflow/internal/models"
	"quickflow/internal/usecase"
	http2 "quickflow/utils/http"
	"strconv"
)

type PostHandler struct {
	postUseCase PostUseCase
}

// NewPostHandler creates new post handler.
func NewPostHandler(postUseCase PostUseCase) *PostHandler {
	return &PostHandler{
		postUseCase: postUseCase,
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
func (p *PostHandler) AddPost(w http.ResponseWriter, r *http.Request) {
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
	isRepostString := r.FormValue("is_repost")

	if len(isRepostString) != 0 {
		postForm.IsRepost, err = strconv.ParseBool(r.FormValue("is_repost"))
		if err != nil {
			http2.WriteJSONError(w, "Failed to parse form", http.StatusBadRequest)
			return
		}
	}

	postForm.Images, err = http2.GetFiles(r, "pics")
	if err != nil {
		http2.WriteJSONError(w, "Failed to get files", http.StatusBadRequest)
		return
	}

	post := postForm.ToPostModel(user.Id)

	err = p.postUseCase.AddPost(r.Context(), post)
	if err != nil {
		http2.WriteJSONError(w, "Failed to add post", http.StatusInternalServerError)
		return
	}
}

// DeletePost удаляет пост
// @Summary Удалить пост
// @Description Удаляет пост из ленты
// @Tags Feed
// @Param post_id path string true "Идентификатор поста"
// @Success 200 {string} string "OK"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 403 {object} forms.ErrorForm "Пост не принадлежит пользователю"
// @Failure 404 {object} forms.ErrorForm "Пост не найден"
// @Failure 500 {object} forms.ErrorForm "Ошибка сервера"
// @Router /api/post/{post_id} [delete]
func (p *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	// extracting user from context
	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

	postIdString := mux.Vars(r)["post_id"]
	if len(postIdString) == 0 {
		http2.WriteJSONError(w, "Failed to get post id", http.StatusBadRequest)
		return
	}

	postId, err := uuid.Parse(postIdString)
	if err != nil {
		http2.WriteJSONError(w, "Failed to parse post id", http.StatusBadRequest)
		return
	}

	err = p.postUseCase.DeletePost(r.Context(), user, postId)
	if errors.Is(err, usecase.ErrPostDoesNotBelongToUser) {
		http2.WriteJSONError(w, "Post does not belong to user", http.StatusForbidden)
		return
	} else if errors.Is(err, usecase.ErrPostNotFound) {
		http2.WriteJSONError(w, "Post not found", http.StatusNotFound)
		return
	} else {
		http2.WriteJSONError(w, "Failed delete post", http.StatusInternalServerError)
		return
	}
}
