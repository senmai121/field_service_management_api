package handlers

import (
	"encoding/json"
	"net/http"

	"field_service_management_api/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ListPriorityLevels(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(r.Context(),
			`SELECT id, name, display_order, response_hours, resolve_hours, color_hex, created_at
			 FROM fsm.priority_levels ORDER BY display_order ASC`)
		if err != nil {
			writeError(w, 500, "query failed")
			return
		}
		defer rows.Close()
		items := []models.PriorityLevel{}
		for rows.Next() {
			var item models.PriorityLevel
			if err := rows.Scan(&item.ID, &item.Name, &item.DisplayOrder, &item.ResponseHours, &item.ResolveHours, &item.ColorHex, &item.CreatedAt); err != nil {
				writeError(w, 500, "scan failed")
				return
			}
			items = append(items, item)
		}
		writeJSON(w, 200, models.ListResponse[models.PriorityLevel]{Data: items, Total: len(items)})
	}
}

func CreatePriorityLevel(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.PriorityLevelRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "invalid request body")
			return
		}
		if req.Name == "" {
			writeError(w, 400, "name is required")
			return
		}
		var item models.PriorityLevel
		err := db.QueryRow(r.Context(),
			`INSERT INTO fsm.priority_levels (name, display_order, response_hours, resolve_hours, color_hex)
			 VALUES ($1, $2, $3, $4, $5)
			 RETURNING id, name, display_order, response_hours, resolve_hours, color_hex, created_at`,
			req.Name, req.DisplayOrder, req.ResponseHours, req.ResolveHours, req.ColorHex,
		).Scan(&item.ID, &item.Name, &item.DisplayOrder, &item.ResponseHours, &item.ResolveHours, &item.ColorHex, &item.CreatedAt)
		if err != nil {
			writeError(w, 500, "insert failed")
			return
		}
		writeJSON(w, 201, item)
	}
}

func UpdatePriorityLevel(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var req models.PriorityLevelRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "invalid request body")
			return
		}
		if req.Name == "" {
			writeError(w, 400, "name is required")
			return
		}
		var item models.PriorityLevel
		err := db.QueryRow(r.Context(),
			`UPDATE fsm.priority_levels
			 SET name=$1, display_order=$2, response_hours=$3, resolve_hours=$4, color_hex=$5
			 WHERE id=$6
			 RETURNING id, name, display_order, response_hours, resolve_hours, color_hex, created_at`,
			req.Name, req.DisplayOrder, req.ResponseHours, req.ResolveHours, req.ColorHex, id,
		).Scan(&item.ID, &item.Name, &item.DisplayOrder, &item.ResponseHours, &item.ResolveHours, &item.ColorHex, &item.CreatedAt)
		if err != nil {
			writeError(w, 500, "update failed")
			return
		}
		writeJSON(w, 200, item)
	}
}

func DeletePriorityLevel(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		_, err := db.Exec(r.Context(), `DELETE FROM fsm.priority_levels WHERE id=$1`, id)
		if err != nil {
			writeError(w, 500, "delete failed")
			return
		}
		writeJSON(w, 200, map[string]string{"message": "deleted"})
	}
}
