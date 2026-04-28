package handlers

import (
	"encoding/json"
	"net/http"

	"field_service_management_api/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ListCustomerSites(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(r.Context(),
			`SELECT id, customer_id, name, address, latitude, longitude, is_active, created_at, updated_at
			 FROM fsm.customer_sites ORDER BY created_at DESC`)
		if err != nil {
			writeError(w, 500, "query failed")
			return
		}
		defer rows.Close()
		items := []models.CustomerSite{}
		for rows.Next() {
			var item models.CustomerSite
			if err := rows.Scan(&item.ID, &item.CustomerID, &item.Name, &item.Address, &item.Latitude, &item.Longitude, &item.IsActive, &item.CreatedAt, &item.UpdatedAt); err != nil {
				writeError(w, 500, "scan failed")
				return
			}
			items = append(items, item)
		}
		writeJSON(w, 200, models.ListResponse[models.CustomerSite]{Data: items, Total: len(items)})
	}
}

func CreateCustomerSite(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CustomerSiteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "invalid request body")
			return
		}
		if req.CustomerID == "" || req.Name == "" {
			writeError(w, 400, "customer_id and name are required")
			return
		}
		var item models.CustomerSite
		err := db.QueryRow(r.Context(),
			`INSERT INTO fsm.customer_sites (customer_id, name, address, latitude, longitude, is_active)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 RETURNING id, customer_id, name, address, latitude, longitude, is_active, created_at, updated_at`,
			req.CustomerID, req.Name, req.Address, req.Latitude, req.Longitude, req.IsActive,
		).Scan(&item.ID, &item.CustomerID, &item.Name, &item.Address, &item.Latitude, &item.Longitude, &item.IsActive, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			writeError(w, 500, "insert failed")
			return
		}
		writeJSON(w, 201, item)
	}
}

func UpdateCustomerSite(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var req models.CustomerSiteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "invalid request body")
			return
		}
		if req.Name == "" {
			writeError(w, 400, "name is required")
			return
		}
		var item models.CustomerSite
		err := db.QueryRow(r.Context(),
			`UPDATE fsm.customer_sites
			 SET customer_id=$1, name=$2, address=$3, latitude=$4, longitude=$5, is_active=$6, updated_at=NOW()
			 WHERE id=$7
			 RETURNING id, customer_id, name, address, latitude, longitude, is_active, created_at, updated_at`,
			req.CustomerID, req.Name, req.Address, req.Latitude, req.Longitude, req.IsActive, id,
		).Scan(&item.ID, &item.CustomerID, &item.Name, &item.Address, &item.Latitude, &item.Longitude, &item.IsActive, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			writeError(w, 500, "update failed")
			return
		}
		writeJSON(w, 200, item)
	}
}

func DeleteCustomerSite(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		_, err := db.Exec(r.Context(), `DELETE FROM fsm.customer_sites WHERE id=$1`, id)
		if err != nil {
			writeError(w, 500, "delete failed")
			return
		}
		writeJSON(w, 200, map[string]string{"message": "deleted"})
	}
}
