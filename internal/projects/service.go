// internal/projects/service.go
package projects

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Project struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Title          string    `json:"title"`
	Shape          string    `json:"shape"`
	SizeInches     float64   `json:"size_inches"`
	NailCount      int       `json:"nail_count"`
	NailStyle      string    `json:"nail_style"`
	NailDiameterMM float64   `json:"nail_diameter_mm"`
	LayerMode      bool      `json:"layer_mode"`
	LayerCount     int       `json:"layer_count"`
	ImageRemoteURL string    `json:"image_remote_url"`
	StringPlanJSON string    `json:"string_plan_json"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ListMeta struct {
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service { return &Service{db: db} }

func (s *Service) List(ctx context.Context, userID string, page, limit int) ([]Project, ListMeta, error) {
	offset := (page - 1) * limit
	var total int
	s.db.QueryRow(ctx, `SELECT COUNT(*) FROM projects WHERE user_id = $1`, userID).Scan(&total)

	rows, err := s.db.Query(ctx,
		`SELECT id, user_id, title, shape, size_inches, nail_count, nail_style,
		        nail_diameter_mm, layer_mode, layer_count, image_remote_url,
		        string_plan_json, status, created_at, updated_at
		 FROM projects WHERE user_id = $1
		 ORDER BY updated_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, ListMeta{}, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		rows.Scan(&p.ID, &p.UserID, &p.Title, &p.Shape, &p.SizeInches,
			&p.NailCount, &p.NailStyle, &p.NailDiameterMM, &p.LayerMode,
			&p.LayerCount, &p.ImageRemoteURL, &p.StringPlanJSON, &p.Status,
			&p.CreatedAt, &p.UpdatedAt)
		projects = append(projects, p)
	}
	if projects == nil {
		projects = []Project{}
	}
	return projects, ListMeta{Total: total, Page: page, Limit: limit}, nil
}

func (s *Service) Create(ctx context.Context, userID string, body map[string]any) (*Project, error) {
	p := &Project{
		ID:        uuid.New().String(),
		UserID:    userID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if v, ok := body["title"].(string); ok {
		p.Title = v
	}
	if v, ok := body["shape"].(string); ok {
		p.Shape = v
	}
	if v, ok := body["size_inches"].(float64); ok {
		p.SizeInches = v
	}
	if v, ok := body["nail_count"].(float64); ok {
		p.NailCount = int(v)
	}
	if v, ok := body["nail_style"].(string); ok {
		p.NailStyle = v
	}
	if v, ok := body["nail_diameter_mm"].(float64); ok {
		p.NailDiameterMM = v
	}
	if v, ok := body["layer_mode"].(bool); ok {
		p.LayerMode = v
	}
	if v, ok := body["layer_count"].(float64); ok {
		p.LayerCount = int(v)
	}
	if v, ok := body["image_remote_url"].(string); ok {
		p.ImageRemoteURL = v
	}
	if v, ok := body["string_plan_json"].(string); ok {
		p.StringPlanJSON = v
	}
	if v, ok := body["status"].(string); ok {
		p.Status = v
	}

	_, err := s.db.Exec(ctx,
		`INSERT INTO projects (id, user_id, title, shape, size_inches, nail_count,
		  nail_style, nail_diameter_mm, layer_mode, layer_count, image_remote_url,
		  string_plan_json, status, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		p.ID, p.UserID, p.Title, p.Shape, p.SizeInches, p.NailCount,
		p.NailStyle, p.NailDiameterMM, p.LayerMode, p.LayerCount,
		p.ImageRemoteURL, p.StringPlanJSON, p.Status, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert project: %w", err)
	}
	return p, nil
}

func (s *Service) GetByID(ctx context.Context, id, userID string) (*Project, error) {
	p := &Project{}
	err := s.db.QueryRow(ctx,
		`SELECT id, user_id, title, shape, size_inches, nail_count, nail_style,
		        nail_diameter_mm, layer_mode, layer_count, image_remote_url,
		        string_plan_json, status, created_at, updated_at
		 FROM projects WHERE id = $1 AND user_id = $2`, id, userID,
	).Scan(&p.ID, &p.UserID, &p.Title, &p.Shape, &p.SizeInches,
		&p.NailCount, &p.NailStyle, &p.NailDiameterMM, &p.LayerMode,
		&p.LayerCount, &p.ImageRemoteURL, &p.StringPlanJSON, &p.Status,
		&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("project not found")
	}
	return p, nil
}

func (s *Service) Update(ctx context.Context, id, userID string, body map[string]any) (*Project, error) {
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	i := 1

	fields := map[string]string{
		"title": "title", "shape": "shape", "status": "status",
		"image_remote_url": "image_remote_url", "string_plan_json": "string_plan_json",
		"nail_style": "nail_style",
	}
	for key, col := range fields {
		if v, ok := body[key].(string); ok {
			sets = append(sets, fmt.Sprintf("%s = $%d", col, i))
			args = append(args, v)
			i++
		}
	}
	numFields := map[string]string{
		"size_inches": "size_inches", "nail_diameter_mm": "nail_diameter_mm",
	}
	for key, col := range numFields {
		if v, ok := body[key].(float64); ok {
			sets = append(sets, fmt.Sprintf("%s = $%d", col, i))
			args = append(args, v)
			i++
		}
	}

	args = append(args, id, userID)
	query := fmt.Sprintf(
		`UPDATE projects SET %s WHERE id = $%d AND user_id = $%d`,
		strings.Join(sets, ", "), i, i+1,
	)
	s.db.Exec(ctx, query, args...)
	return s.GetByID(ctx, id, userID)
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	_, err := s.db.Exec(ctx,
		`DELETE FROM projects WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (s *Service) Export(ctx context.Context, id, userID, format string) (string, error) {
	p, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return "", err
	}
	switch format {
	case "json":
		return p.StringPlanJSON, nil
	case "txt":
		return buildTXT(p), nil
	default:
		return p.StringPlanJSON, nil
	}
}

func buildTXT(p *Project) string {
	var sb strings.Builder
	sb.WriteString("ThreadCraft Instructions\n")
	sb.WriteString(fmt.Sprintf("Project: %s\n", p.Title))
	sb.WriteString(fmt.Sprintf("Shape: %s | Size: %.0f\" | Nails: %d | Layers: %d\n",
		strings.ToUpper(p.Shape), p.SizeInches, p.NailCount, p.LayerCount))
	sb.WriteString(fmt.Sprintf("Mounting: %s (%.1fmm diameter)\n",
		strings.ToUpper(strings.ReplaceAll(p.NailStyle, "_", " ")), p.NailDiameterMM))
	sb.WriteString("================================================\n")
	sb.WriteString("\n(Full step-by-step list requires string_plan_json parsing)\n")
	return sb.String()
}
