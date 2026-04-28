package models

import (
	"fmt"
	"strings"
	"time"
)

// FlexTime accepts datetime-local ("2006-01-02T15:04") and RFC3339 in JSON.
type FlexTime struct{ time.Time }

func (ft *FlexTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "null" || s == "" {
		return nil
	}
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			ft.Time = t
			return nil
		}
	}
	return fmt.Errorf("cannot parse time %q", s)
}

// ---- Reference tables ----

type ServiceType struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}
type ServiceTypeRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	IsActive    bool    `json:"is_active"`
}

type PriorityLevel struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	DisplayOrder  int       `json:"display_order"`
	ResponseHours *int      `json:"response_hours,omitempty"`
	ResolveHours  *int      `json:"resolve_hours,omitempty"`
	ColorHex      *string   `json:"color_hex,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
type PriorityLevelRequest struct {
	Name          string  `json:"name"`
	DisplayOrder  int     `json:"display_order"`
	ResponseHours *int    `json:"response_hours,omitempty"`
	ResolveHours  *int    `json:"resolve_hours,omitempty"`
	ColorHex      *string `json:"color_hex,omitempty"`
}

type SkillCategory struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
type SkillCategoryRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type AssetCategory struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
type AssetCategoryRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// ---- Customer ----

type Customer struct {
	ID        string    `json:"id"`
	Code      *string   `json:"code,omitempty"`
	Name      string    `json:"name"`
	TaxID     *string   `json:"tax_id,omitempty"`
	Email     *string   `json:"email,omitempty"`
	Phone     *string   `json:"phone,omitempty"`
	Address   *string   `json:"address,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type CustomerRequest struct {
	Code     *string `json:"code,omitempty"`
	Name     string  `json:"name"`
	TaxID    *string `json:"tax_id,omitempty"`
	Email    *string `json:"email,omitempty"`
	Phone    *string `json:"phone,omitempty"`
	Address  *string `json:"address,omitempty"`
	IsActive bool    `json:"is_active"`
}

type CustomerSite struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	Name       string    `json:"name"`
	Address    *string   `json:"address,omitempty"`
	Latitude   *float64  `json:"latitude,omitempty"`
	Longitude  *float64  `json:"longitude,omitempty"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
type CustomerSiteRequest struct {
	CustomerID string   `json:"customer_id"`
	Name       string   `json:"name"`
	Address    *string  `json:"address,omitempty"`
	Latitude   *float64 `json:"latitude,omitempty"`
	Longitude  *float64 `json:"longitude,omitempty"`
	IsActive   bool     `json:"is_active"`
}

// ---- Technician ----

type Technician struct {
	ID        string    `json:"id"`
	UserID    *int64    `json:"user_id,omitempty"`
	Username  *string   `json:"username,omitempty"`
	Code      *string   `json:"code,omitempty"`
	FullName  string    `json:"full_name"`
	Phone     *string   `json:"phone,omitempty"`
	Email     *string   `json:"email,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type TechnicianRequest struct {
	UserID   *int64  `json:"user_id,omitempty"`
	Code     *string `json:"code,omitempty"`
	FullName string  `json:"full_name"`
	Phone    *string `json:"phone,omitempty"`
	Email    *string `json:"email,omitempty"`
	IsActive bool    `json:"is_active"`
}

// UserItem — for listing system users (used in technician assignment dropdown)
type UserItem struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// ---- Work Order ----

type WorkOrder struct {
	ID              string     `json:"id"`
	OrderNo         *string    `json:"order_no,omitempty"`
	CustomerID      string     `json:"customer_id"`
	CustomerName    string     `json:"customer_name"`
	CustomerSiteID  string     `json:"customer_site_id"`
	SiteName        string     `json:"site_name"`
	AssetID         *string    `json:"asset_id,omitempty"`
	AssetName       *string    `json:"asset_name,omitempty"`
	AssetSerial     *string    `json:"asset_serial,omitempty"`
	ServiceTypeID   *string    `json:"service_type_id,omitempty"`
	ServiceTypeName *string    `json:"service_type_name,omitempty"`
	PriorityLevelID *string    `json:"priority_level_id,omitempty"`
	PriorityName    *string    `json:"priority_name,omitempty"`
	PriorityColor   *string    `json:"priority_color,omitempty"`
	Status          string     `json:"status"`
	Title           string     `json:"title"`
	Description     *string    `json:"description,omitempty"`
	ScheduledStart  *time.Time `json:"scheduled_start,omitempty"`
	ScheduledEnd    *time.Time `json:"scheduled_end,omitempty"`
	ActualStart     *time.Time `json:"actual_start,omitempty"`
	ActualEnd       *time.Time `json:"actual_end,omitempty"`
	SLADueAt        *time.Time `json:"sla_due_at,omitempty"`
	RepairCost      *float64   `json:"repair_cost,omitempty"`
	WarrantyCovered bool       `json:"warranty_covered"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
type WorkOrderRequest struct {
	CustomerID      string    `json:"customer_id"`
	CustomerSiteID  string    `json:"customer_site_id"`
	AssetID         *string   `json:"asset_id,omitempty"`
	ServiceTypeID   *string   `json:"service_type_id,omitempty"`
	PriorityLevelID *string   `json:"priority_level_id,omitempty"`
	Status          string    `json:"status"`
	Title           string    `json:"title"`
	Description     *string   `json:"description,omitempty"`
	ScheduledStart  *FlexTime `json:"scheduled_start,omitempty"`
	ScheduledEnd    *FlexTime `json:"scheduled_end,omitempty"`
	RepairCost      *float64  `json:"repair_cost,omitempty"`
	WarrantyCovered bool      `json:"warranty_covered"`
}
type WorkOrderStatusRequest struct {
	Status      string    `json:"status"`
	ActualStart *FlexTime `json:"actual_start,omitempty"`
	ActualEnd   *FlexTime `json:"actual_end,omitempty"`
}

// ---- Work Order Assignment ----

type Assignment struct {
	TechnicianID string    `json:"technician_id"`
	FullName     string    `json:"full_name"`
	Code         *string   `json:"code,omitempty"`
	Phone        *string   `json:"phone,omitempty"`
	IsLead       bool      `json:"is_lead"`
	AssignedAt   time.Time `json:"assigned_at"`
}

type AssignRequest struct {
	TechnicianID string `json:"technician_id"`
	IsLead       bool   `json:"is_lead"`
}

type ListResponse[T any] struct {
	Data  []T `json:"data"`
	Total int `json:"total"`
}

// ---- Asset ----

type Asset struct {
	ID                string     `json:"id"`
	CustomerSiteID    string     `json:"customer_site_id"`
	SiteName          string     `json:"site_name"`
	CustomerName      string     `json:"customer_name"`
	AssetCategoryID   *string    `json:"asset_category_id,omitempty"`
	CategoryName      *string    `json:"category_name,omitempty"`
	SerialNo          *string    `json:"serial_no,omitempty"`
	Name              string     `json:"name"`
	Brand             *string    `json:"brand,omitempty"`
	Model             *string    `json:"model,omitempty"`
	InstalledAt       *time.Time `json:"installed_at,omitempty"`
	WarrantyExpiresAt *time.Time `json:"warranty_expires_at,omitempty"`
	Status            string     `json:"status"`
	Notes             *string    `json:"notes,omitempty"`
	Latitude          *float64   `json:"latitude,omitempty"`
	Longitude         *float64   `json:"longitude,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type AssetRequest struct {
	CustomerSiteID    string    `json:"customer_site_id"`
	AssetCategoryID   *string   `json:"asset_category_id,omitempty"`
	SerialNo          *string   `json:"serial_no,omitempty"`
	Name              string    `json:"name"`
	Brand             *string   `json:"brand,omitempty"`
	Model             *string   `json:"model,omitempty"`
	InstalledAt       *FlexTime `json:"installed_at,omitempty"`
	WarrantyExpiresAt *FlexTime `json:"warranty_expires_at,omitempty"`
	Status            string    `json:"status"`
	Notes             *string   `json:"notes,omitempty"`
	Latitude          *float64  `json:"latitude,omitempty"`
	Longitude         *float64  `json:"longitude,omitempty"`
}

// ---- Technician My Work Orders ----

type MyWorkOrder struct {
	ID              string     `json:"id"`
	OrderNo         *string    `json:"order_no,omitempty"`
	Title           string     `json:"title"`
	Description     *string    `json:"description,omitempty"`
	Status          string     `json:"status"`
	CustomerName    string     `json:"customer_name"`
	SiteName        string     `json:"site_name"`
	SiteAddress     *string    `json:"site_address,omitempty"`
	AssetName       *string    `json:"asset_name,omitempty"`
	AssetSerial     *string    `json:"asset_serial,omitempty"`
	ServiceTypeName *string    `json:"service_type_name,omitempty"`
	PriorityName    *string    `json:"priority_name,omitempty"`
	PriorityColor   *string    `json:"priority_color,omitempty"`
	ScheduledStart  *time.Time `json:"scheduled_start,omitempty"`
	ScheduledEnd    *time.Time `json:"scheduled_end,omitempty"`
	ActualStart     *time.Time `json:"actual_start,omitempty"`
	ActualEnd       *time.Time `json:"actual_end,omitempty"`
	SLADueAt        *time.Time `json:"sla_due_at,omitempty"`
	AssignedAt      time.Time  `json:"assigned_at"`
}
