package handlers

import (
	"encoding/json"
	"net/http"

	"field_service_management_api/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ListAssignments GET /api/fsm/work-orders/{id}/assignments
func ListAssignments(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		woID := chi.URLParam(r, "id")
		rows, err := db.Query(r.Context(), `
			SELECT a.technician_id, t.full_name, t.code, t.phone, a.is_lead, a.assigned_at
			FROM fsm.assignments a
			JOIN fsm.technicians t ON t.id = a.technician_id
			WHERE a.work_order_id = $1
			ORDER BY a.is_lead DESC, a.assigned_at ASC`, woID)
		if err != nil {
			writeError(w, 500, "query failed")
			return
		}
		defer rows.Close()
		items := []models.Assignment{}
		for rows.Next() {
			var a models.Assignment
			if err := rows.Scan(&a.TechnicianID, &a.FullName, &a.Code, &a.Phone, &a.IsLead, &a.AssignedAt); err != nil {
				writeError(w, 500, "scan failed")
				return
			}
			items = append(items, a)
		}
		writeJSON(w, 200, models.ListResponse[models.Assignment]{Data: items, Total: len(items)})
	}
}

// AddAssignment POST /api/fsm/work-orders/{id}/assignments
func AddAssignment(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		woID := chi.URLParam(r, "id")
		var req models.AssignRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "invalid request body")
			return
		}
		if req.TechnicianID == "" {
			writeError(w, 400, "technician_id is required")
			return
		}

		// If is_lead, demote any existing lead first
		if req.IsLead {
			if _, err := db.Exec(r.Context(),
				`UPDATE fsm.assignments SET is_lead = false WHERE work_order_id = $1`, woID,
			); err != nil {
				writeError(w, 500, "demote lead failed")
				return
			}
		}

		_, err := db.Exec(r.Context(), `
			INSERT INTO fsm.assignments (work_order_id, technician_id, is_lead)
			VALUES ($1, $2, $3)
			ON CONFLICT (work_order_id, technician_id)
			DO UPDATE SET is_lead = EXCLUDED.is_lead, assigned_at = NOW()`,
			woID, req.TechnicianID, req.IsLead,
		)
		if err != nil {
			writeError(w, 500, "assign failed: "+err.Error())
			return
		}

		// Auto-update work order status to 'assigned' if still draft
		_, _ = db.Exec(r.Context(),
			`UPDATE fsm.work_orders SET status='assigned', updated_at=NOW() WHERE id=$1 AND status='draft'`,
			woID,
		)

		writeJSON(w, 200, map[string]string{"message": "assigned"})
	}
}

// RemoveAssignment DELETE /api/fsm/work-orders/{id}/assignments/{technicianId}
func RemoveAssignment(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		woID := chi.URLParam(r, "id")
		techID := chi.URLParam(r, "technicianId")
		if _, err := db.Exec(r.Context(),
			`DELETE FROM fsm.assignments WHERE work_order_id=$1 AND technician_id=$2`,
			woID, techID,
		); err != nil {
			writeError(w, 500, "remove failed")
			return
		}
		writeJSON(w, 200, map[string]string{"message": "removed"})
	}
}
