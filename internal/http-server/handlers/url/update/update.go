package update

import (
	"log/slog"
	"net/http"

	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Alias  string `json:"alias" validate:"required"`
	NewURL string `json:"new_url" validate:"required,url"`
}

type Response struct {
	resp.Response
	CountUodated int64 `json:"countUpdated"`
}

type URLUpdater interface {
	UpdateURL(alias string, newURL string) (int64, error)
}

func New(log *slog.Logger, urlUpdater URLUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.update.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err = validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("invalid request", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		alias := req.Alias
		url := req.NewURL

		rows, err := urlUpdater.UpdateURL(alias, url)
		if err != nil {
			log.Error("failed to update url", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal server error"))

			return
		}

		render.JSON(w, r, Response{
			Response:     resp.OK(),
			CountUodated: rows,
		})
	}
}
