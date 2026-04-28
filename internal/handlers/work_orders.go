package handlers

import (
	"encoding/json"
	"field_service_management_api/internal/models"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var validWorkOrderStatuses = map[string]bool{
	"draft":       true,
	"assigned":    true,
	"in_progress": true,
	"on_hold":     true,
	"completed":   true,
	"closed":      true,
	"cancelled":   true,
}

const workOrderSelectQuery = `
SELECT
  wo.id, wo.order_no, wo.customer_id, c.name as customer_name,
  wo.customer_site_id, cs.name as site_name,
  wo.asset_id, a.name as asset_name, a.serial_no as asset_serial,
  wo.service_type_id, st.name as service_type_name,
  wo.priority_level_id, pl.name as priority_name, pl.color_hex as priority_color,
  wo.status, wo.title, wo.description,
  wo.scheduled_start, wo.scheduled_end, wo.actual_start, wo.actual_end, wo.sla_due_at,
  wo.repair_cost, wo.warranty_covered,
  wo.created_at, wo.updated_at
FROM fsm.work_orders wo
JOIN fsm.customers c ON c.id = wo.customer_id
JOIN fsm.customer_sites cs ON cs.id = wo.customer_site_id
LEFT JOIN fsm.assets a ON a.id = wo.asset_id
LEFT JOIN fsm.service_types st ON st.id = wo.service_type_id
LEFT JOIN fsm.priority_levels pl ON pl.id = wo.priority_level_id`

func scanWorkOrder(rows interface {
	Scan(...any) error
}, item *models.WorkOrder) error {
	return rows.Scan(
		&item.ID, &item.OrderNo, &item.CustomerID, &item.CustomerName,
		&item.CustomerSiteID, &item.SiteName,
		&item.AssetID, &item.AssetName, &item.AssetSerial,
		&item.ServiceTypeID, &item.ServiceTypeName,
		&item.PriorityLevelID, &item.PriorityName, &item.PriorityColor,
		&item.Status, &item.Title, &item.Description,
		&item.ScheduledStart, &item.ScheduledEnd, &item.ActualStart, &item.ActualEnd, &item.SLADueAt,
		&item.RepairCost, &item.WarrantyCovered,
		&item.CreatedAt, &item.UpdatedAt,
	)
}

func ListWorkOrders(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(r.Context(), workOrderSelectQuery+` ORDER BY wo.created_at DESC`)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer rows.Close()
		items := []models.WorkOrder{}
		for rows.Next() {
			var item models.WorkOrder
			if err := scanWorkOrder(rows, &item); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			items = append(items, item)
		}
		writeJSON(w, 200, models.ListResponse[models.WorkOrder]{Data: items, Total: len(items)})
	}
}

func GetWorkOrder(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var item models.WorkOrder
		row := db.QueryRow(r.Context(), workOrderSelectQuery+` WHERE wo.id=$1`, id)
		if err := scanWorkOrder(row, &item); err != nil {
			writeError(w, 404, "work order not found")
			return
		}
		writeJSON(w, 200, item)
	}
}

func CreateWorkOrder(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.WorkOrderRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "invalid request body")
			return
		}
		if req.CustomerID == "" || req.CustomerSiteID == "" || req.Title == "" {
			writeError(w, 400, "customer_id, customer_site_id, and title are required")
			return
		}
		if req.Status == "" {
			req.Status = "draft"
		}
		if !validWorkOrderStatuses[req.Status] {
			writeError(w, 400, "invalid status")
			return
		}

		var scheduledStart, scheduledEnd *time.Time
		if req.ScheduledStart != nil {
			scheduledStart = &req.ScheduledStart.Time
		}
		if req.ScheduledEnd != nil {
			scheduledEnd = &req.ScheduledEnd.Time
		}

		var id, orderNo string
		err := db.QueryRow(r.Context(),
			`INSERT INTO fsm.work_orders (order_no, customer_id, customer_site_id, asset_id, service_type_id, priority_level_id, status, title, description, scheduled_start, scheduled_end, repair_cost, warranty_covered, created_by)
			 VALUES (
			   'WO-' || TO_CHAR(NOW(), 'YYYYMMDD') || '-' || LPAD((SELECT COUNT(*)+1 FROM fsm.work_orders)::text, 4, '0'),
			   $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NULL
			 )
			 RETURNING id, order_no`,
			req.CustomerID, req.CustomerSiteID, req.AssetID, req.ServiceTypeID, req.PriorityLevelID,
			req.Status, req.Title, req.Description, scheduledStart, scheduledEnd,
			req.RepairCost, req.WarrantyCovered,
		).Scan(&id, &orderNo)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Return full work order with JOINs
		var item models.WorkOrder
		row := db.QueryRow(r.Context(), workOrderSelectQuery+` WHERE wo.id=$1`, id)
		if err := scanWorkOrder(row, &item); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, 201, item)
	}
}

func UpdateWorkOrder(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var req models.WorkOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "invalid request body")
			return
		}
		if req.CustomerID == "" || req.CustomerSiteID == "" || req.Title == "" {
			writeError(w, 400, "customer_id, customer_site_id, and title are required")
			return
		}
		if req.Status != "" && !validWorkOrderStatuses[req.Status] {
			writeError(w, 400, "invalid status")
			return
		}

		var scheduledStart, scheduledEnd *time.Time
		if req.ScheduledStart != nil {
			scheduledStart = &req.ScheduledStart.Time
		}
		if req.ScheduledEnd != nil {
			scheduledEnd = &req.ScheduledEnd.Time
		}

		_, err := db.Exec(r.Context(),
			`UPDATE fsm.work_orders
			 SET customer_id=$1, customer_site_id=$2, asset_id=$3, service_type_id=$4, priority_level_id=$5,
			     status=$6, title=$7, description=$8, scheduled_start=$9, scheduled_end=$10,
			     repair_cost=$11, warranty_covered=$12, updated_at=NOW()
			 WHERE id=$13`,
			req.CustomerID, req.CustomerSiteID, req.AssetID, req.ServiceTypeID, req.PriorityLevelID,
			req.Status, req.Title, req.Description, scheduledStart, scheduledEnd,
			req.RepairCost, req.WarrantyCovered, id,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		var item models.WorkOrder
		row := db.QueryRow(r.Context(), workOrderSelectQuery+` WHERE wo.id=$1`, id)
		if err := scanWorkOrder(row, &item); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, 200, item)
	}
}

func UpdateWorkOrderStatus(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var req models.WorkOrderStatusRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "invalid request body")
			return
		}
		if !validWorkOrderStatuses[req.Status] {
			writeError(w, 400, "invalid status; must be one of: draft, assigned, in_progress, on_hold, completed, closed, cancelled")
			return
		}

		var actualStart, actualEnd *time.Time
		if req.ActualStart != nil {
			actualStart = &req.ActualStart.Time
		}
		if req.ActualEnd != nil {
			actualEnd = &req.ActualEnd.Time
		}
		_, err := db.Exec(r.Context(),
			`UPDATE fsm.work_orders
			 SET status=$1,
			     actual_start = COALESCE($2, actual_start),
			     actual_end   = COALESCE($3, actual_end),
			     updated_at   = NOW()
			 WHERE id=$4`,
			req.Status, actualStart, actualEnd, id,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		var item models.WorkOrder
		row := db.QueryRow(r.Context(), workOrderSelectQuery+` WHERE wo.id=$1`, id)
		if err := scanWorkOrder(row, &item); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, 200, item)
	}
}
