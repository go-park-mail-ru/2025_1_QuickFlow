package http

import (
    "errors"
    "fmt"
    "net/http"
    "quickflow/pkg/logger"
    "strconv"

    "github.com/google/uuid"
    "github.com/gorilla/mux"

    "quickflow/internal/delivery/forms"
    "quickflow/internal/models"
    "quickflow/internal/usecase"
    http2 "quickflow/utils/http"
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
    ctx := http2.SetRequestId(r.Context())
    user, ok := ctx.Value("user").(models.User)
    if !ok {
        logger.Error(ctx, "Failed to get user from context while adding post")
        http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
        return
    }
    logger.Info(ctx, fmt.Sprintf("User %s requested to add post", user.Login))

    err := r.ParseMultipartForm(10 << 20) // 10 MB
    if err != nil {
        logger.Error(ctx, fmt.Sprintf("Failed to parse form: %s", err.Error()))
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
            logger.Error(ctx, fmt.Sprintf("Failed to parse is_repost: %s", err.Error()))
            http2.WriteJSONError(w, "Failed to parse form", http.StatusBadRequest)
            return
        }
    }

    postForm.Images, err = http2.GetFiles(r, "pics")
    if err != nil {
        logger.Error(ctx, fmt.Sprintf("Failed to get files: %s", err.Error()))
        http2.WriteJSONError(w, "Failed to get files", http.StatusBadRequest)
        return
    }
    logger.Info(ctx, fmt.Sprintf("Recieved post: %+v", postForm))

    post := postForm.ToPostModel(user.Id)

    err = p.postUseCase.AddPost(ctx, post)
    if err != nil {
        logger.Error(ctx, fmt.Sprintf("Failed to add post: %s", err.Error()))
        http2.WriteJSONError(w, "Failed to add post", http.StatusInternalServerError)
        return
    }
    logger.Info(ctx, "Successfully added post")
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
    ctx := http2.SetRequestId(r.Context())
    user, ok := ctx.Value("user").(models.User)
    if !ok {
        logger.Error(ctx, "Failed to get user from context while deleting post")
        http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
        return
    }

    postIdString := mux.Vars(r)["post_id"]
    if len(postIdString) == 0 {
        logger.Error(ctx, "Failed to get post id from request")
        http2.WriteJSONError(w, "Failed to get post id", http.StatusBadRequest)
        return
    }
    logger.Info(ctx, fmt.Sprintf("User %s requested to delete post with id %s", user.Login, postIdString))

    postId, err := uuid.Parse(postIdString)
    if err != nil {
        logger.Error(ctx, fmt.Sprintf("Failed to parse post id: %s", err.Error()))
        http2.WriteJSONError(w, "Failed to parse post id", http.StatusBadRequest)
        return
    }

    err = p.postUseCase.DeletePost(ctx, user, postId)
    if errors.Is(err, usecase.ErrPostDoesNotBelongToUser) {
        logger.Error(ctx, fmt.Sprintf("Post %s does not belong to user %s", postIdString, user.Login))
        http2.WriteJSONError(w, "Post does not belong to user", http.StatusForbidden)
        return
    } else if errors.Is(err, usecase.ErrPostNotFound) {
        logger.Error(ctx, fmt.Sprintf("Post %s not found", postIdString))
        http2.WriteJSONError(w, "Post not found", http.StatusNotFound)
        return
    } else if err != nil {
        logger.Error(ctx, fmt.Sprintf("Failed to delete post: %s", err.Error()))
        http2.WriteJSONError(w, "Failed delete post", http.StatusInternalServerError)
        return
    }
    logger.Info(ctx, fmt.Sprintf("Successfully deleted post %s", postIdString))
}
