package handlers

import (
	"net/http"

	"field_service_management_api/internal/middleware"
	"field_service_management_api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

func GetMyWorkOrders(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.GetClaims(r)
		if claims == nil {
			writeError(w, 401, "unauthorized")
			return
		}

		var technicianID string
		err := db.QueryRow(r.Context(),
			`SELECT id FROM fsm.technicians WHERE user_id = $1 AND is_active = true LIMIT 1`,
			claims.UserID,
		).Scan(&technicianID)
		if err != nil {
			writeError(w, 404, "no technician profile linked to this user")
			return
		}

		rows, err := db.Query(r.Context(), `
			SELECT
			  wo.id, wo.order_no, wo.title, wo.description, wo.status,
			  c.name,
			  cs.name, cs.address,
			  a.name, a.serial_no,
			  st.name,
			  pl.name, pl.color_hex,
			  wo.scheduled_start, wo.scheduled_end,
			  wo.actual_start, wo.actual_end,
			  wo.sla_due_at,
			  asgn.assigned_at
			FROM fsm.assignments asgn
			JOIN fsm.work_orders wo ON wo.id = asgn.work_order_id
			JOIN fsm.customers c ON c.id = wo.customer_id
			JOIN fsm.customer_sites cs ON cs.id = wo.customer_site_id
			LEFT JOIN fsm.assets a ON a.id = wo.asset_id
			LEFT JOIN fsm.service_types st ON st.id = wo.service_type_id
			LEFT JOIN fsm.priority_levels pl ON pl.id = wo.priority_level_id
			WHERE asgn.technician_id = $1
			  AND wo.status NOT IN ('closed','cancelled')
			ORDER BY
			  CASE wo.status
			    WHEN 'in_progress' THEN 1
			    WHEN 'assigned'    THEN 2
			    WHEN 'on_hold'     THEN 3
			    ELSE 4
			  END,
			  wo.scheduled_start ASC NULLS LAST`,
			technicianID,
		)
		if err != nil {
			writeError(w, 500, "query failed")
			return
		}
		defer rows.Close()

		items := []models.MyWorkOrder{}
		for rows.Next() {
			var m models.MyWorkOrder
			if err := rows.Scan(
				&m.ID, &m.OrderNo, &m.Title, &m.Description, &m.Status,
				&m.CustomerName,
				&m.SiteName, &m.SiteAddress,
				&m.AssetName, &m.AssetSerial,
				&m.ServiceTypeName,
				&m.PriorityName, &m.PriorityColor,
				&m.ScheduledStart, &m.ScheduledEnd,
				&m.ActualStart, &m.ActualEnd,
				&m.SLADueAt,
				&m.AssignedAt,
			); err != nil {
				writeError(w, 500, "scan failed")
				return
			}
			items = append(items, m)
		}
		writeJSON(w, 200, models.ListResponse[models.MyWorkOrder]{Data: items, Total: len(items)})
	}
}
