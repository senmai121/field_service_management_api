package handlers

import (
	"encoding/json"
	"net/http"

	"field_service_management_api/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ListServiceTypes(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(r.Context(),
			`SELECT id, name, description, is_active, created_at FROM fsm.service_types ORDER BY created_at DESC`)
		if err != nil {
			writeError(w, 500, "query failed")
			return
		}
		defer rows.Close()
		items := []models.ServiceType{}
		for rows.Next() {
			var item models.ServiceType
			if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.IsActive, &item.CreatedAt); err != nil {
				writeError(w, 500, "scan failed")
				return
			}
			items = append(items, item)
		}
		writeJSON(w, 200, models.ListResponse[models.ServiceType]{Data: items, Total: len(items)})
	}
}

func CreateServiceType(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.ServiceTypeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "invalid request body")
			return
		}
		if req.Name == "" {
			writeError(w, 400, "name is required")
			return
		}
		var item models.ServiceType
		err := db.QueryRow(r.Context(),
			`INSERT INTO fsm.service_types (name, description, is_active)
			 VALUES ($1, $2, $3)
			 RETURNING id, name, description, is_active, created_at`,
			req.Name, req.Description, req.IsActive,
		).Scan(&item.ID, &item.Name, &item.Description, &item.IsActive, &item.CreatedAt)
		if err != nil {
			writeError(w, 500, "insert failed")
			return
		}
		writeJSON(w, 201, item)
	}
}

func UpdateServiceType(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var req models.ServiceTypeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "invalid request body")
			return
		}
		if req.Name == "" {
			writeError(w, 400, "name is required")
			return
		}
		var item models.ServiceType
		err := db.QueryRow(r.Context(),
			`UPDATE fsm.service_types SET name=$1, description=$2, is_active=$3
			 WHERE id=$4
			 RETURNING id, name, description, is_active, created_at`,
			req.Name, req.Description, req.IsActive, id,
		).Scan(&item.ID, &item.Name, &item.Description, &item.IsActive, &item.CreatedAt)
		if err != nil {
			writeError(w, 500, "update failed")
			return
		}
		writeJSON(w, 200, item)
	}
}

func DeleteServiceType(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		_, err := db.Exec(r.Context(), `DELETE FROM fsm.service_types WHERE id=$1`, id)
		if err != nil {
			writeError(w, 500, "delete failed")
			return
		}
		writeJSON(w, 200, map[string]string{"message": "deleted"})
	}
}
