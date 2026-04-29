package handlers

import (
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// WorkOrderStatusItem represents a single status bucket.
type WorkOrderStatusItem struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// OnTimeItem represents one month's on-time vs late counts.
type OnTimeItem struct {
	Month   string `json:"month"` // "2025-01"
	OnTime  int    `json:"on_time"`
	Late    int    `json:"late"`
}

// TechnicianHoursItem represents hours worked by one technician this month.
type TechnicianHoursItem struct {
	TechnicianID   string  `json:"technician_id"`
	TechnicianName string  `json:"technician_name"`
	Hours          float64 `json:"hours"`
}

// GetWorkOrderStatus returns COUNT per status for all work orders.
func GetWorkOrderStatus(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := pool.Query(r.Context(), `
			SELECT status, COUNT(*)::int AS count
			FROM fsm.work_orders
			GROUP BY status
			ORDER BY status
		`)
		if err != nil {
			log.Printf("[GetWorkOrderStatus] query error: %v", err)
			writeError(w, http.StatusInternalServerError, "query failed")
			return
		}
		defer rows.Close()

		var items []WorkOrderStatusItem
		for rows.Next() {
			var item WorkOrderStatusItem
			if err := rows.Scan(&item.Status, &item.Count); err != nil {
				log.Printf("[GetWorkOrderStatus] scan error: %v", err)
				writeError(w, http.StatusInternalServerError, "scan failed")
				return
			}
			items = append(items, item)
		}
		if items == nil {
			items = []WorkOrderStatusItem{}
		}
		writeJSON(w, http.StatusOK, items)
	}
}

// GetOnTimeCompletion returns on-time vs late completed work orders for the last 6 months.
// "On-time" = actual_end <= scheduled_end; "Late" = actual_end > scheduled_end or not completed but deadline passed.
func GetOnTimeCompletion(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := pool.Query(r.Context(), `
			WITH months AS (
				SELECT generate_series(
					date_trunc('month', NOW() - INTERVAL '5 months'),
					date_trunc('month', NOW()),
					'1 month'
				) AS month_start
			),
			classified AS (
				SELECT
					date_trunc('month', COALESCE(actual_end, scheduled_end)) AS month_start,
					CASE
						WHEN status = 'completed' AND actual_end IS NOT NULL AND actual_end <= scheduled_end THEN 'on_time'
						ELSE 'late'
					END AS result
				FROM fsm.work_orders
				WHERE scheduled_end IS NOT NULL
				  AND (
				  		(status = 'completed' AND actual_end IS NOT NULL)
				  		OR (status != 'completed' AND scheduled_end < NOW())
				  )
			)
			SELECT
				to_char(m.month_start, 'YYYY-MM') AS month,
				COALESCE(SUM(CASE WHEN c.result = 'on_time' THEN 1 ELSE 0 END), 0)::int AS on_time,
				COALESCE(SUM(CASE WHEN c.result = 'late'    THEN 1 ELSE 0 END), 0)::int AS late
			FROM months m
			LEFT JOIN classified c ON c.month_start = m.month_start
			GROUP BY m.month_start
			ORDER BY m.month_start
		`)
		if err != nil {
			log.Printf("[GetOnTimeCompletion] query error: %v", err)
			writeError(w, http.StatusInternalServerError, "query failed")
			return
		}
		defer rows.Close()

		var items []OnTimeItem
		for rows.Next() {
			var item OnTimeItem
			if err := rows.Scan(&item.Month, &item.OnTime, &item.Late); err != nil {
				log.Printf("[GetOnTimeCompletion] scan error: %v", err)
				writeError(w, http.StatusInternalServerError, "scan failed")
				return
			}
			items = append(items, item)
		}
		if items == nil {
			items = []OnTimeItem{}
		}
		writeJSON(w, http.StatusOK, items)
	}
}

// GetTechnicianHours returns total actual working hours per technician for the current month.
// Falls back to scheduled_start / scheduled_end when actual values are missing,
// so that completed work orders with partial time data are still counted.
func GetTechnicianHours(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := pool.Query(r.Context(), `
			SELECT
				t.id,
				t.full_name,
				COALESCE(
					SUM(
						EXTRACT(EPOCH FROM (
							COALESCE(wo.actual_end,   wo.scheduled_end)
							- COALESCE(wo.actual_start, wo.scheduled_start)
						)) / 3600.0
					),
					0
				) AS hours
			FROM fsm.technicians t
			LEFT JOIN fsm.assignments woa ON woa.technician_id = t.id
			LEFT JOIN fsm.work_orders wo
				ON wo.id = woa.work_order_id
				AND wo.status = 'completed'
				AND (wo.actual_end IS NOT NULL OR wo.scheduled_end IS NOT NULL)
				AND (wo.actual_start IS NOT NULL OR wo.scheduled_start IS NOT NULL)
				AND COALESCE(wo.actual_end, wo.scheduled_end) > COALESCE(wo.actual_start, wo.scheduled_start)
				AND date_trunc('month', COALESCE(wo.actual_start, wo.scheduled_start)) = date_trunc('month', NOW())
			GROUP BY t.id, t.full_name
			ORDER BY hours DESC, t.full_name
		`)
		if err != nil {
			log.Printf("[GetTechnicianHours] query error: %v", err)
			writeError(w, http.StatusInternalServerError, "query failed")
			return
		}
		defer rows.Close()

		var items []TechnicianHoursItem
		for rows.Next() {
			var item TechnicianHoursItem
			if err := rows.Scan(&item.TechnicianID, &item.TechnicianName, &item.Hours); err != nil {
				log.Printf("[GetTechnicianHours] scan error: %v", err)
				writeError(w, http.StatusInternalServerError, "scan failed")
				return
			}
			// Round to 1 decimal
			item.Hours = float64(int(item.Hours*10+0.5)) / 10
			items = append(items, item)
		}
		if items == nil {
			items = []TechnicianHoursItem{}
		}
		writeJSON(w, http.StatusOK, items)
	}
}
