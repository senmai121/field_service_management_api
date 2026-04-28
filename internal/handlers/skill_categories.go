package handlers

import (
	"encoding/json"
	"net/http"

	"field_service_management_api/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ListSkillCategories(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(r.Context(),
			`SELECT id, name, description, created_at FROM fsm.skill_categories ORDER BY created_at DESC`)
		if err != nil {
			writeError(w, 500, "query failed")
			return
		}
		defer rows.Close()
		items := []models.SkillCategory{}
		for rows.Next() {
			var item models.SkillCategory
			if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt); err != nil {
				writeError(w, 500, "scan failed")
				return
			}
			items = append(items, item)
		}
		writeJSON(w, 200, models.ListResponse[models.SkillCategory]{Data: items, Total: len(items)})
	}
}

func CreateSkillCategory(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.SkillCategoryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "invalid request body")
			return
		}
		if req.Name == "" {
			writeError(w, 400, "name is required")
			return
		}
		var item models.SkillCategory
		err := db.QueryRow(r.Context(),
			`INSERT INTO fsm.skill_categories (name, description)
			 VALUES ($1, $2)
			 RETURNING id, name, description, created_at`,
			req.Name, req.Description,
		).Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt)
		if err != nil {
			writeError(w, 500, "insert failed")
			return
		}
		writeJSON(w, 201, item)
	}
}

func UpdateSkillCategory(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var req models.SkillCategoryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "invalid request body")
			return
		}
		if req.Name == "" {
			writeError(w, 400, "name is required")
			return
		}
		var item models.SkillCategory
		err := db.QueryRow(r.Context(),
			`UPDATE fsm.skill_categories SET name=$1, description=$2
			 WHERE id=$3
			 RETURNING id, name, description, created_at`,
			req.Name, req.Description, id,
		).Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt)
		if err != nil {
			writeError(w, 500, "update failed")
			return
		}
		writeJSON(w, 200, item)
	}
}

func DeleteSkillCategory(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		_, err := db.Exec(r.Context(), `DELETE FROM fsm.skill_categories WHERE id=$1`, id)
		if err != nil {
			writeError(w, 500, "delete failed")
			return
		}
		writeJSON(w, 200, map[string]string{"message": "deleted"})
	}
}
