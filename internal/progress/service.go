// internal/progress/service.go
package progress

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Marker struct {
	ID        string    `json:"id"`
	Step      int       `json:"step"`
	Label     string    `json:"label"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

type Progress struct {
	ProjectID   string     `json:"project_id"`
	CurrentStep int        `json:"current_step"`
	TotalSteps  int        `json:"total_steps"`
	LastUpdated time.Time  `json:"last_updated"`
	CompletedAt *time.Time `json:"completed_at"`
	Markers     []Marker   `json:"markers"`
}

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service { return &Service{db: db} }

func (s *Service) Get(ctx context.Context, projectID, userID string) (*Progress, error) {
	p := &Progress{ProjectID: projectID, Markers: []Marker{}}

	err := s.db.QueryRow(ctx,
		`SELECT current_step, total_steps, last_updated, completed_at
		 FROM project_progress
		 WHERE project_id = $1 AND user_id = $2`, projectID, userID,
	).Scan(&p.CurrentStep, &p.TotalSteps, &p.LastUpdated, &p.CompletedAt)
	if err != nil {
		// Return empty progress if none exists yet
		p.LastUpdated = time.Now().UTC()
		return p, nil
	}

	rows, err := s.db.Query(ctx,
		`SELECT id, step, label, note, created_at
		 FROM progress_markers WHERE project_id = $1 ORDER BY step`, projectID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var m Marker
			rows.Scan(&m.ID, &m.Step, &m.Label, &m.Note, &m.CreatedAt)
			p.Markers = append(p.Markers, m)
		}
	}
	return p, nil
}

func (s *Service) Put(ctx context.Context, projectID, userID string, p *Progress) (*Progress, error) {
	p.ProjectID = projectID
	p.LastUpdated = time.Now().UTC()

	_, err := s.db.Exec(ctx,
		`INSERT INTO project_progress (project_id, user_id, current_step, total_steps, last_updated)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (project_id) DO UPDATE
		   SET current_step = EXCLUDED.current_step,
		       total_steps  = EXCLUDED.total_steps,
		       last_updated = EXCLUDED.last_updated`,
		projectID, userID, p.CurrentStep, p.TotalSteps, p.LastUpdated,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert progress: %w", err)
	}

	// Replace markers
	s.db.Exec(ctx, `DELETE FROM progress_markers WHERE project_id = $1`, projectID)
	for _, m := range p.Markers {
		if m.ID == "" {
			m.ID = uuid.New().String()
		}
		if m.CreatedAt.IsZero() {
			m.CreatedAt = time.Now().UTC()
		}
		s.db.Exec(ctx,
			`INSERT INTO progress_markers (id, project_id, step, label, note, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			m.ID, projectID, m.Step, m.Label, m.Note, m.CreatedAt,
		)
	}
	return p, nil
}
