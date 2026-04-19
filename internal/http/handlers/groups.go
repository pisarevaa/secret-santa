package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"unicode/utf8"

	"github.com/andreypisarev/secret-santa/internal/db/sqlc"
	"github.com/andreypisarev/secret-santa/internal/groups"
	mw "github.com/andreypisarev/secret-santa/internal/http/middleware"
	"github.com/go-chi/chi/v5"
)

type GroupHandler struct {
	Queries *sqlc.Queries
}

func (h *GroupHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := mw.UserIDFromContext(r.Context())

	var req struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "неверный формат запроса")
		return
	}

	if req.Title == "" || utf8.RuneCountInString(req.Title) > 100 {
		writeError(w, http.StatusBadRequest, "invalid_input", "название группы должно быть от 1 до 100 символов")
		return
	}

	code, err := groups.GenerateInviteCode()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	group, err := h.Queries.CreateGroup(r.Context(), sqlc.CreateGroupParams{
		InviteCode:  code,
		Title:       req.Title,
		OrganizerID: userID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":          group.ID,
		"invite_code": group.InviteCode,
	})
}

func (h *GroupHandler) GetByInviteCode(w http.ResponseWriter, r *http.Request) {
	inviteCode := chi.URLParam(r, "inviteCode")

	group, err := h.Queries.GetGroupByInviteCode(r.Context(), inviteCode)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not_found", "группа не найдена")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	userID, hasSession := mw.UserIDFromContext(r.Context())

	if !hasSession {
		count, _ := h.Queries.CountMembersByGroup(r.Context(), group.ID)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"id":           group.ID,
			"title":        group.Title,
			"member_count": count,
			"status":       group.Status,
		})
		return
	}

	members, err := h.Queries.ListMembershipsByGroup(r.Context(), group.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	isMember := false
	isOrganizer := group.OrganizerID == userID
	var myMembershipID *int64

	memberList := make([]map[string]interface{}, 0, len(members))
	for _, m := range members {
		user, _ := h.Queries.GetUserByID(r.Context(), m.UserID)
		memberList = append(memberList, map[string]interface{}{
			"name":  user.Name,
			"is_me": m.UserID == userID,
		})
		if m.UserID == userID {
			isMember = true
			id := m.ID
			myMembershipID = &id
		}
	}

	if !isMember {
		count, _ := h.Queries.CountMembersByGroup(r.Context(), group.ID)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"id":           group.ID,
			"title":        group.Title,
			"member_count": count,
			"status":       group.Status,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":               group.ID,
		"title":            group.Title,
		"status":           group.Status,
		"members":          memberList,
		"is_organizer":     isOrganizer,
		"my_membership_id": myMembershipID,
	})
}

func (h *GroupHandler) Join(w http.ResponseWriter, r *http.Request) {
	userID, _ := mw.UserIDFromContext(r.Context())
	inviteCode := chi.URLParam(r, "inviteCode")

	var req struct {
		Name     string `json:"name"`
		Wishlist string `json:"wishlist"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "неверный формат запроса")
		return
	}

	if req.Name == "" || utf8.RuneCountInString(req.Name) > 50 {
		writeError(w, http.StatusBadRequest, "invalid_input", "имя должно быть от 1 до 50 символов")
		return
	}
	if utf8.RuneCountInString(req.Wishlist) > 2000 {
		writeError(w, http.StatusBadRequest, "invalid_input", "вишлист не должен превышать 2000 символов")
		return
	}

	group, err := h.Queries.GetGroupByInviteCode(r.Context(), inviteCode)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not_found", "группа не найдена")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	if group.Status != "open" {
		writeError(w, http.StatusConflict, "already_drawn", "жеребьевка уже проведена")
		return
	}

	if err := h.Queries.UpdateUserName(r.Context(), sqlc.UpdateUserNameParams{
		Name: req.Name,
		ID:   userID,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	_, err = h.Queries.CreateMembership(r.Context(), sqlc.CreateMembershipParams{
		GroupID:  group.ID,
		UserID:   userID,
		Wishlist: req.Wishlist,
	})
	if err != nil {
		writeError(w, http.StatusConflict, "already_member", "вы уже участник этой группы")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *GroupHandler) UpdateWishlist(w http.ResponseWriter, r *http.Request) {
	userID, _ := mw.UserIDFromContext(r.Context())
	membershipIDStr := chi.URLParam(r, "id")

	var req struct {
		Wishlist string `json:"wishlist"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "неверный формат запроса")
		return
	}

	if utf8.RuneCountInString(req.Wishlist) > 2000 {
		writeError(w, http.StatusBadRequest, "invalid_input", "вишлист не должен превышать 2000 символов")
		return
	}

	id, err := strconv.ParseInt(membershipIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "некорректный id")
		return
	}

	membership, err := h.Queries.GetMembership(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "участие не найдено")
		return
	}

	if membership.UserID != userID {
		writeError(w, http.StatusForbidden, "forbidden", "нет доступа")
		return
	}

	if err := h.Queries.UpdateWishlist(r.Context(), sqlc.UpdateWishlistParams{
		Wishlist: req.Wishlist,
		ID:       id,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
