package handlers

import (
	"database/sql"
	"errors"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/andreypisarev/secret-santa/internal/db/sqlc"
	"github.com/andreypisarev/secret-santa/internal/draw"
	mw "github.com/andreypisarev/secret-santa/internal/http/middleware"
	"github.com/go-chi/chi/v5"
)

type DrawHandler struct {
	Queries *sqlc.Queries
	DB      *sql.DB
}

func (h *DrawHandler) Draw(w http.ResponseWriter, r *http.Request) {
	userID, _ := mw.UserIDFromContext(r.Context())
	groupIDStr := chi.URLParam(r, "id")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "неверный ID группы")
		return
	}

	group, err := h.Queries.GetGroupByID(r.Context(), groupID)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not_found", "группа не найдена")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	if group.OrganizerID != userID {
		writeError(w, http.StatusForbidden, "forbidden", "только организатор может провести жеребьевку")
		return
	}

	if group.Status != "open" {
		writeError(w, http.StatusConflict, "already_drawn", "жеребьевка уже проведена")
		return
	}

	members, err := h.Queries.ListMembershipsByGroup(r.Context(), groupID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	participantIDs := make([]int64, len(members))
	for i, m := range members {
		participantIDs[i] = m.UserID
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	assignments, err := draw.Assign(participantIDs, rng)
	if errors.Is(err, draw.ErrNotEnoughMembers) {
		writeError(w, http.StatusBadRequest, "not_enough_members", "нужно минимум 2 участника")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}
	defer tx.Rollback()

	qtx := h.Queries.WithTx(tx)

	res, err := qtx.DrawGroup(r.Context(), groupID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		writeError(w, http.StatusConflict, "already_drawn", "жеребьевка уже проведена")
		return
	}

	for santaID, recipientID := range assignments {
		if err := qtx.SetRecipient(r.Context(), sqlc.SetRecipientParams{
			RecipientID: sql.NullInt64{Int64: recipientID, Valid: true},
			GroupID:     groupID,
			UserID:      santaID,
		}); err != nil {
			writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DrawHandler) MyRecipient(w http.ResponseWriter, r *http.Request) {
	userID, _ := mw.UserIDFromContext(r.Context())
	groupIDStr := chi.URLParam(r, "id")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "неверный ID группы")
		return
	}

	recipient, err := h.Queries.GetMyRecipient(r.Context(), sqlc.GetMyRecipientParams{
		GroupID: groupID,
		UserID:  userID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not_found", "подопечный не назначен")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"recipient": map[string]interface{}{
			"name":     recipient.Name,
			"wishlist": recipient.Wishlist,
		},
	})
}
