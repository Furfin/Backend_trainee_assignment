package user

import (
	"encoding/csv"
	"errors"
	"example/ravito/initializers"
	"example/ravito/models"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
)

type RequestUserSlugs struct {
	AddTo      []string       `json:"AddTo,omitempty"`
	RemoveFrom []string       `json:"RemoveFrom,omitempty"`
	Exparation map[string]int `json:"ttl_days,omitempty"`
}

type RequestUser struct {
	Userid int64 `json:"userid" validate:"required"`
}

type RequestUserStory struct {
	Year  int64 `json:"year" validate:"required"`
	Month int64 `json:"month" validate:"required"`
}

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type UserResponse struct {
	Status   string `json:"status"`
	Error    string `json:"error"`
	Segments []models.Segment
}

// UserSegmentsUpdate - Creates segment using unique slug from body
// @description If needed, creates user, then adds user to specified segments and deletes from specified
// @description In optional dictionary you can specify expiration date for every added segments
// @Tags ravito
// @Accept  json
// @Produce  json
// @Param userid path int true "user id"
// @Param request body user.RequestUserSlugs true "query params"
// @Success 200 {object} user.Response "api response"
// @Router /user/{userid}/add [post]
func UserSegmentsUpdate(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.user.add.New"
	log := initializers.Log
	log = log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	userid, err := strconv.Atoi(chi.URLParam(r, "userid"))

	if err != nil || userid < 0 {
		log.Error("Invalid id in req params: " + err.Error())
		render.Status(r, 400)
		render.JSON(w, r, Response{"Error", "Invalid userid"})

		return
	}

	var req RequestUserSlugs

	err = render.DecodeJSON(r.Body, &req)
	if errors.Is(err, io.EOF) {
		log.Error("request body is empty")

		render.Status(r, 400)
		render.JSON(w, r, Response{"Error", "Empty request"})

		return
	}
	if err != nil {
		log.Error("failed to decode request body")

		render.Status(r, 400)
		render.JSON(w, r, Response{"Error", "Failed to decode request" + ": " + err.Error()})

		return
	}

	log.Info("request body decoded", slog.Any("request", req))
	v := validator.New()

	if err := v.Struct(req); err != nil {
		validateErr := err.(validator.ValidationErrors)

		log.Error("invalid request")

		render.Status(r, 400)
		render.JSON(w, r, Response{"Error", validateErr.Error()})

		return
	}

	// fmt.Println(req.AddTo[0])
	var user models.User

	if initializers.DB.First(&user, "userid = ?", userid).Error != nil {
		user = models.User{Userid: int64(userid)}
		result := initializers.DB.Create(&user)
		if result.Error != nil {
			log.Error("invalid request")
			render.Status(r, 400)
			render.JSON(w, r, Response{"Error", "Db creation problem"})

			return
		}
	}

	var seg models.Segment
	for _, val := range req.AddTo {
		seg = models.Segment{}
		if initializers.DB.Where(&seg, "slug = ?", val).Error != nil {
			log.Error("invalid request")
			render.Status(r, 400)
			render.JSON(w, r, Response{"Error", "Invalid segment slug: " + val})

			return
		}
	}
	for _, val := range req.RemoveFrom {
		seg = models.Segment{}
		if initializers.DB.Where(&seg, "slug = ?", val).Error != nil {
			log.Error("invalid request")
			render.Status(r, 400)
			render.JSON(w, r, Response{"Error", "Invalid segment slug: " + val})

			return
		}
	}
	var rel models.UserSegment
	for _, val := range req.AddTo {
		seg = models.Segment{}
		rel = models.UserSegment{}
		initializers.DB.First(&seg, "slug = ?", val)
		if initializers.DB.Where("user_id = ?", user.ID).First(&rel, "segment_id = ?", seg.ID).Error == nil {
			log.Error("invalid request")
			render.Status(r, 400)
			render.JSON(w, r, Response{"Error", strconv.FormatInt(user.Userid, 10) + " already in " + val})

			return
		}
		rel = models.UserSegment{UserID: int(user.ID), User: user, SegmentID: int(seg.ID), Segment: seg}
		expire, ex := req.Exparation[seg.Slug]
		if ex && expire < 0 {
			log.Error("invalid request")
			render.Status(r, 400)
			render.JSON(w, r, Response{"Error", "Invalid data in ttl map"})
			return
		} else if ex {
			rel.DaysExpire = expire
		}

		result := initializers.DB.Create(&rel)
		if result.Error != nil {
			log.Error("invalid request")
			render.Status(r, 400)
			render.JSON(w, r, Response{"Error", "db creation problem"})

			return
		}
	}

	for _, val := range req.RemoveFrom {
		seg = models.Segment{}
		rel = models.UserSegment{}
		initializers.DB.First(&seg, "slug = ?", val)
		if initializers.DB.Where("user_id = ?", user.ID).First(&rel, "segment_id = ?", seg.ID).Error != nil {
			log.Error("invalid request")
			render.Status(r, 400)
			render.JSON(w, r, Response{"Error", strconv.FormatInt(user.Userid, 10) + " not in " + val})

			return
		}

		result := initializers.DB.Delete(&rel)
		if result.Error != nil {
			log.Error("invalid request")
			render.Status(r, 400)
			render.JSON(w, r, Response{"Error", "Db creation problem"})

			return
		}
	}

	render.JSON(w, r, Response{"Ok", ""})
}

// UserGetInfo - Returns segments of user
// @Tags ravito
// @Accept  json
// @Produce  json
// @Param userid path int true "user id"
// @Success 200 {object} user.UserResponse "api response"
// @Router /user/{userid} [get]
func GetUserInfo(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.user.add.NewUser"
	log := initializers.Log
	log = log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	userid, err := strconv.Atoi(chi.URLParam(r, "userid"))

	if err != nil || userid < 0 {
		log.Error("Invalid id in req params")
		render.Status(r, 400)
		render.JSON(w, r, Response{"Error", "Invalid userid"})

		return
	}

	var user models.User
	var seg models.Segment
	var segs []models.Segment
	var rels []models.UserSegment

	if initializers.DB.First(&user, "userid = ?", userid).Error != nil {
		user = models.User{Userid: int64(userid)}
		result := initializers.DB.Create(&user)
		if result.Error != nil {
			log.Error("invalid request")
			render.Status(r, 400)
			render.JSON(w, r, Response{"Error", "Db creation problem"})

			return
		}
	}

	initializers.DB.Where("user_id = ?", user.ID).Find(&rels)
	fmt.Println(rels, "ahhahahha")
	for _, rel := range rels {
		seg = models.Segment{}
		if initializers.DB.First(&seg, "id = ?", rel.SegmentID).Error == nil {
			segs = append(segs, seg)
		}
	}

	render.JSON(w, r, UserResponse{"Ok", "", segs})
}

// UserGetHistory - Returns user segments addition deletion history
// @description year and month params set the left border of time interval
// @description in which search will be conducted
// @Tags ravito
// @Accept  json
// @Produce  json
// @Param userid path int true "user id"
// @Param request body user.RequestUserStory true "query params"
// @Success 200 {file} binary
// @Router /user/{userid}/csv [post]
func GetUserHistory(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.user.add.NewUser"
	log := initializers.Log
	log = log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var req RequestUserStory
	userid, err := strconv.Atoi(chi.URLParam(r, "userid"))

	if err != nil || userid < 0 {
		log.Error("Invalid id in req params")
		render.Status(r, 400)
		render.JSON(w, r, Response{"Error", "Invalid userid"})

		return
	}

	err = render.DecodeJSON(r.Body, &req)
	if errors.Is(err, io.EOF) {
		log.Error("request body is empty")
		render.Status(r, 400)
		render.JSON(w, r, Response{"Error", "Empty request"})

		return
	}
	if err != nil {
		log.Error("failed to decode request body")
		render.Status(r, 400)
		render.JSON(w, r, Response{"Error", "Failed to decode request" + ": " + err.Error()})

		return
	}

	log.Info("request body decoded", slog.Any("request", req))
	v := validator.New()

	if err := v.Struct(req); err != nil {
		validateErr := err.(validator.ValidationErrors)

		log.Error("invalid request")
		render.Status(r, 400)
		render.JSON(w, r, Response{"Error", validateErr.Error()})

		return
	}

	// fmt.Println(req.AddTo[0])
	var user models.User
	var rels []models.UserSegment

	if initializers.DB.First(&user, "userid = ?", userid).Error != nil {
		user = models.User{Userid: int64(userid)}
		result := initializers.DB.Create(&user)
		if result.Error != nil {
			log.Error("invalid request")
			render.Status(r, 400)
			render.JSON(w, r, Response{"Error", "Db creation problem"})

			return
		}
	}
	if req.Year < 0 || req.Month < 0 || req.Month > 12 {
		log.Error("invalid request")
		render.Status(r, 400)
		render.JSON(w, r, Response{"Error", "Invalid year/month data"})
		return
	}
	records := [][]string{{"user", "segment_slug", "added/deleted", "datetime"}}
	initializers.DB.Unscoped().Where("user_id = ?", user.ID).Find(&rels)
	for _, rel := range rels {
		seg := models.Segment{}
		initializers.DB.Where("id = ?", rel.SegmentID).Find(&seg)
		if rel.CreatedAt.Year() >= int(req.Year) && rel.CreatedAt.Month() >= time.Month(req.Month) {
			records = append(records, []string{strconv.FormatInt(user.Userid, 10), seg.Slug, "added", rel.CreatedAt.String()})
		}

	}
	rels = []models.UserSegment{}

	initializers.DB.Unscoped().Where("user_id = ?", user.ID).Where("deleted_at IS NOT NULL").Find(&rels)
	for _, rel := range rels {
		seg := models.Segment{}
		initializers.DB.Where("id = ?", rel.SegmentID).Find(&seg)
		if rel.DeletedAt.Time.Year() >= int(req.Year) && rel.DeletedAt.Time.Month() >= time.Month(req.Month) {
			records = append(records, []string{strconv.FormatInt(user.Userid, 10), seg.Slug, "deleted", rel.DeletedAt.Time.String()})
		}
	}
	wr := csv.NewWriter(w)
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Add("Content-Disposition", `attachment; filename="history.csv"`)
	if err := wr.WriteAll(records); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
