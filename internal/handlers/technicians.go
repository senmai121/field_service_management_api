package handlers

import (
	"encoding/json"
	"net/http"

	"field_service_management_api/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const techSelectQuery = `
SELECT t.id, t.user_id, u.username, t.code, t.full_name, t.phone, t.email, t.is_active, t.created_at, t.updated_at
FROM fsm.technicians t
LEFT JOIN fsm.users u ON u.id = t.user_id`

func scanTechnician(row interface{ Scan(...any) error }) (models.Technician, error) {
	var item models.Technician
	err := row.Scan(&item.ID, &item.UserID, &item.Username, &item.Code, &item.FullName, &item.Phone, &item.Email, &item.IsActive, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func ListTechnicians(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(r.Context(), techSelectQuery+` ORDER BY t.created_at DESC`)
		if err != nil {
			writeError(w, 500, "query failed")
			return
		}
		defer rows.Close()
		items := []models.Technician{}
		for rows.Next() {
			item, err := scanTechnician(rows)
			if err != nil {
				writeError(w, 500, "scan failed")
				return
			}
			items = append(items, item)
		}
		writeJSON(w, 200, models.ListResponse[models.Technician]{Data: items, Total: len(items)})
	}
}

func GetTechnician(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		item, err := scanTechnician(db.QueryRow(r.Context(), techSelectQuery+` WHERE t.id=$1`, id))
		if err != nil {
			writeError(w, 404, "technician not found")
			return
		}
		writeJSON(w, 200, item)
	}
}

func CreateTechnician(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.TechnicianRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "invalid request body")
			return
		}
		if req.FullName == "" {
			writeError(w, 400, "full_name is required")
			return
		}
		var id string
		err := db.QueryRow(r.Context(),
			`INSERT INTO fsm.technicians (user_id, code, full_name, phone, email, is_active)
			 VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
			req.UserID, req.Code, req.FullName, req.Phone, req.Email, req.IsActive,
		).Scan(&id)
		if err != nil {
			writeError(w, 500, "insert failed: "+err.Error())
			return
		}
		item, err := scanTechnician(db.QueryRow(r.Context(), techSelectQuery+` WHERE t.id=$1`, id))
		if err != nil {
			writeError(w, 500, "fetch after insert failed")
			return
		}
		writeJSON(w, 201, item)
	}
}

func UpdateTechnician(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var req models.TechnicianRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "invalid request body")
			return
		}
		if req.FullName == "" {
			writeError(w, 400, "full_name is required")
			return
		}
		_, err := db.Exec(r.Context(),
			`UPDATE fsm.technicians
			 SET user_id=$1, code=$2, full_name=$3, phone=$4, email=$5, is_active=$6, updated_at=NOW()
			 WHERE id=$7`,
			req.UserID, req.Code, req.FullName, req.Phone, req.Email, req.IsActive, id,
		)
		if err != nil {
			writeError(w, 500, "update failed: "+err.Error())
			return
		}
		item, err := scanTechnician(db.QueryRow(r.Context(), techSelectQuery+` WHERE t.id=$1`, id))
		if err != nil {
			writeError(w, 404, "technician not found after update")
			return
		}
		writeJSON(w, 200, item)
	}
}

func DeleteTechnician(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		_, err := db.Exec(r.Context(), `DELETE FROM fsm.technicians WHERE id=$1`, id)
		if err != nil {
			writeError(w, 500, "delete failed")
			return
		}
		writeJSON(w, 200, map[string]string{"message": "deleted"})
	}
}

func ListUsers(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(r.Context(), `SELECT id, username, email FROM fsm.users ORDER BY username`)
		if err != nil {
			writeError(w, 500, "query failed")
			return
		}
		defer rows.Close()
		items := []models.UserItem{}
		for rows.Next() {
			var u models.UserItem
			if err := rows.Scan(&u.ID, &u.Username, &u.Email); err != nil {
				writeError(w, 500, "scan failed")
				return
			}
			items = append(items, u)
		}
		writeJSON(w, 200, models.ListResponse[models.UserItem]{Data: items, Total: len(items)})
	}
}
