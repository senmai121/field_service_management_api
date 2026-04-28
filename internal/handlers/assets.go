package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"field_service_management_api/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const assetSelectQuery = `
SELECT
  a.id, a.customer_site_id, cs.name, c.name,
  a.asset_category_id, ac.name,
  a.serial_no, a.name, a.brand, a.model,
  a.installed_at, a.warranty_expires_at, a.status, a.notes,
  a.latitude, a.longitude,
  a.created_at, a.updated_at
FROM fsm.assets a
JOIN fsm.customer_sites cs ON cs.id = a.customer_site_id
JOIN fsm.customers c ON c.id = cs.customer_id
LEFT JOIN fsm.asset_categories ac ON ac.id = a.asset_category_id`

type assetScanner interface {
	Scan(...any) error
}

func scanAsset(row assetScanner) (models.Asset, error) {
	var a models.Asset
	err := row.Scan(
		&a.ID, &a.CustomerSiteID, &a.SiteName, &a.CustomerName,
		&a.AssetCategoryID, &a.CategoryName,
		&a.SerialNo, &a.Name, &a.Brand, &a.Model,
		&a.InstalledAt, &a.WarrantyExpiresAt, &a.Status, &a.Notes,
		&a.Latitude, &a.Longitude,
		&a.CreatedAt, &a.UpdatedAt,
	)
	return a, err
}

func ListAssets(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(r.Context(), assetSelectQuery+` ORDER BY a.name`)
		if err != nil {
			writeError(w, 500, "query failed")
			return
		}
		defer rows.Close()
		items := []models.Asset{}
		for rows.Next() {
			a, err := scanAsset(rows)
			if err != nil {
				writeError(w, 500, "scan failed")
				return
			}
			items = append(items, a)
		}
		writeJSON(w, 200, models.ListResponse[models.Asset]{Data: items, Total: len(items)})
	}
}

func ListAssetsBySite(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		siteID := chi.URLParam(r, "id")
		rows, err := db.Query(r.Context(),
			assetSelectQuery+` WHERE a.customer_site_id = $1 AND a.status = 'active' ORDER BY a.name`,
			siteID,
		)
		if err != nil {
			writeError(w, 500, "query failed")
			return
		}
		defer rows.Close()
		items := []models.Asset{}
		for rows.Next() {
			a, err := scanAsset(rows)
			if err != nil {
				writeError(w, 500, "scan failed")
				return
			}
			items = append(items, a)
		}
		writeJSON(w, 200, models.ListResponse[models.Asset]{Data: items, Total: len(items)})
	}
}

func GetAsset(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		a, err := scanAsset(db.QueryRow(r.Context(), assetSelectQuery+` WHERE a.id = $1`, id))
		if err != nil {
			writeError(w, 404, "asset not found")
			return
		}
		writeJSON(w, 200, a)
	}
}

func CreateAsset(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.AssetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "invalid request body")
			return
		}
		if req.Name == "" || req.CustomerSiteID == "" {
			writeError(w, 400, "name and customer_site_id are required")
			return
		}
		if req.Status == "" {
			req.Status = "active"
		}
		var installedAt, warrantyExpiresAt *time.Time
		if req.InstalledAt != nil {
			installedAt = &req.InstalledAt.Time
		}
		if req.WarrantyExpiresAt != nil {
			warrantyExpiresAt = &req.WarrantyExpiresAt.Time
		}
		var id string
		err := db.QueryRow(r.Context(), `
			INSERT INTO fsm.assets
			  (customer_site_id, asset_category_id, serial_no, name, brand, model,
			   installed_at, warranty_expires_at, status, notes, latitude, longitude)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
			RETURNING id`,
			req.CustomerSiteID, req.AssetCategoryID, req.SerialNo, req.Name,
			req.Brand, req.Model, installedAt, warrantyExpiresAt,
			req.Status, req.Notes, req.Latitude, req.Longitude,
		).Scan(&id)
		if err != nil {
			writeError(w, 500, "insert failed: "+err.Error())
			return
		}
		a, err := scanAsset(db.QueryRow(r.Context(), assetSelectQuery+` WHERE a.id = $1`, id))
		if err != nil {
			writeError(w, 500, "fetch after insert failed")
			return
		}
		writeJSON(w, 201, a)
	}
}

func UpdateAsset(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var req models.AssetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "invalid request body")
			return
		}
		var installedAt, warrantyExpiresAt *time.Time
		if req.InstalledAt != nil {
			installedAt = &req.InstalledAt.Time
		}
		if req.WarrantyExpiresAt != nil {
			warrantyExpiresAt = &req.WarrantyExpiresAt.Time
		}
		_, err := db.Exec(r.Context(), `
			UPDATE fsm.assets SET
			  customer_site_id=$1, asset_category_id=$2, serial_no=$3, name=$4,
			  brand=$5, model=$6, installed_at=$7, warranty_expires_at=$8,
			  status=$9, notes=$10, latitude=$11, longitude=$12, updated_at=NOW()
			WHERE id=$13`,
			req.CustomerSiteID, req.AssetCategoryID, req.SerialNo, req.Name,
			req.Brand, req.Model, installedAt, warrantyExpiresAt,
			req.Status, req.Notes, req.Latitude, req.Longitude, id,
		)
		if err != nil {
			writeError(w, 500, "update failed: "+err.Error())
			return
		}
		a, err := scanAsset(db.QueryRow(r.Context(), assetSelectQuery+` WHERE a.id = $1`, id))
		if err != nil {
			writeError(w, 404, "asset not found after update")
			return
		}
		writeJSON(w, 200, a)
	}
}

func DeleteAsset(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, err := db.Exec(r.Context(), `DELETE FROM fsm.assets WHERE id = $1`, id); err != nil {
			writeError(w, 500, "delete failed")
			return
		}
		writeJSON(w, 200, map[string]string{"message": "deleted"})
	}
}
