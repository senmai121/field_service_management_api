package handlers

import (
	"encoding/json"
	"net/http"

	"field_service_management_api/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ListCustomers(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(r.Context(),
			`SELECT id, code, name, tax_id, email, phone, address, is_active, created_at, updated_at
			 FROM fsm.customers ORDER BY created_at DESC`)
		if err != nil {
			writeError(w, 500, "query failed")
			return
		}
		defer rows.Close()
		items := []models.Customer{}
		for rows.Next() {
			var item models.Customer
			if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.TaxID, &item.Email, &item.Phone, &item.Address, &item.IsActive, &item.CreatedAt, &item.UpdatedAt); err != nil {
				writeError(w, 500, "scan failed")
				return
			}
			items = append(items, item)
		}
		writeJSON(w, 200, models.ListResponse[models.Customer]{Data: items, Total: len(items)})
	}
}

func GetCustomer(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var item models.Customer
		err := db.QueryRow(r.Context(),
			`SELECT id, code, name, tax_id, email, phone, address, is_active, created_at, updated_at
			 FROM fsm.customers WHERE id=$1`, id,
		).Scan(&item.ID, &item.Code, &item.Name, &item.TaxID, &item.Email, &item.Phone, &item.Address, &item.IsActive, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			writeError(w, 404, "customer not found")
			return
		}
		writeJSON(w, 200, item)
	}
}

func CreateCustomer(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CustomerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "invalid request body")
			return
		}
		if req.Name == "" {
			writeError(w, 400, "name is required")
			return
		}
		var item models.Customer
		err := db.QueryRow(r.Context(),
			`INSERT INTO fsm.customers (code, name, tax_id, email, phone, address, is_active)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)
			 RETURNING id, code, name, tax_id, email, phone, address, is_active, created_at, updated_at`,
			req.Code, req.Name, req.TaxID, req.Email, req.Phone, req.Address, req.IsActive,
		).Scan(&item.ID, &item.Code, &item.Name, &item.TaxID, &item.Email, &item.Phone, &item.Address, &item.IsActive, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			writeError(w, 500, "insert failed")
			return
		}
		writeJSON(w, 201, item)
	}
}

func UpdateCustomer(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var req models.CustomerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "invalid request body")
			return
		}
		if req.Name == "" {
			writeError(w, 400, "name is required")
			return
		}
		var item models.Customer
		err := db.QueryRow(r.Context(),
			`UPDATE fsm.customers
			 SET code=$1, name=$2, tax_id=$3, email=$4, phone=$5, address=$6, is_active=$7, updated_at=NOW()
			 WHERE id=$8
			 RETURNING id, code, name, tax_id, email, phone, address, is_active, created_at, updated_at`,
			req.Code, req.Name, req.TaxID, req.Email, req.Phone, req.Address, req.IsActive, id,
		).Scan(&item.ID, &item.Code, &item.Name, &item.TaxID, &item.Email, &item.Phone, &item.Address, &item.IsActive, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			writeError(w, 500, "update failed")
			return
		}
		writeJSON(w, 200, item)
	}
}

func DeleteCustomer(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		_, err := db.Exec(r.Context(), `DELETE FROM fsm.customers WHERE id=$1`, id)
		if err != nil {
			writeError(w, 500, "delete failed")
			return
		}
		writeJSON(w, 200, map[string]string{"message": "deleted"})
	}
}

// ListCustomerSitesByCustomer handles GET /api/fsm/customers/{id}/sites
func ListCustomerSitesByCustomer(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		customerID := chi.URLParam(r, "id")
		rows, err := db.Query(r.Context(),
			`SELECT id, customer_id, name, address, latitude, longitude, is_active, created_at, updated_at
			 FROM fsm.customer_sites WHERE customer_id=$1 ORDER BY name`, customerID)
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
