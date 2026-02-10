package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gladiator/ent"
	"gladiator/ent/notebook"
	"gladiator/internal/database"
)

const notebookCachePrefix = "notebook:"
const notebookCacheTTL = 15 * time.Minute

type NotebookService struct {
	ent   *ent.Client
	redis *database.RedisClient
}

func NewNotebookService(entClient *ent.Client, redis *database.RedisClient) *NotebookService {
	return &NotebookService{ent: entClient, redis: redis}
}

type CreateNotebookRequest struct {
	Title         string        `json:"title"`
	Description   string        `json:"description"`
	InitialCells  []interface{}  `json:"initial_cells"`
}

func defaultCells() map[string]interface{} {
	return map[string]interface{}{
		"cells": []interface{}{
			map[string]interface{}{
				"id":         uuid.New().String(),
				"type":       "markdown",
				"content":   "# Welcome\n\nStart writing here.",
				"output":    nil,
				"executed_at": nil,
				"order":     float64(0),
			},
			map[string]interface{}{
				"id":         uuid.New().String(),
				"type":       "code",
				"content":   "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}",
				"output":    nil,
				"executed_at": nil,
				"order":     float64(1),
			},
		},
	}
}

func (s *NotebookService) Create(ctx context.Context, ownerID uuid.UUID, req CreateNotebookRequest) (*ent.Notebook, error) {
	title := req.Title
	if title == "" {
		title = "Untitled"
	}
	if len(title) > 255 {
		title = title[:255]
	}
	desc := req.Description
	if len(desc) > 1000 {
		desc = desc[:1000]
	}
	content := defaultCells()
	if len(req.InitialCells) > 0 {
		content["cells"] = req.InitialCells
	}
	create := s.ent.Notebook.Create().
		SetOwnerID(ownerID).
		SetTitle(title).
		SetContent(content)
	if desc != "" {
		create = create.SetDescription(desc)
	}
	return create.Save(ctx)
}

type ListNotebooksResult struct {
	Notebooks  []NotebookMeta `json:"notebooks"`
	Pagination Pagination     `json:"pagination"`
}

type NotebookMeta struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	IsPublic    bool    `json:"is_public"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	CellCount   int     `json:"cell_count"`
}

type Pagination struct {
	Page        int `json:"page"`
	Limit       int `json:"limit"`
	Total       int `json:"total"`
	TotalPages  int `json:"total_pages"`
}

func (s *NotebookService) List(ctx context.Context, ownerID uuid.UUID, page, limit int, sort, order, search string) (*ListNotebooksResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	q := s.ent.Notebook.Query().Where(notebook.OwnerIDEQ(ownerID))
	if search != "" {
		q = q.Where(notebook.Or(
			notebook.TitleContainsFold(search),
			notebook.DescriptionContainsFold(search),
		))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}
	switch sort {
	case "updated_at", "updated":
		if order == "asc" {
			q = q.Order(ent.Asc(notebook.FieldUpdatedAt))
		} else {
			q = q.Order(ent.Desc(notebook.FieldUpdatedAt))
		}
	case "created_at", "created":
		if order == "asc" {
			q = q.Order(ent.Asc(notebook.FieldCreatedAt))
		} else {
			q = q.Order(ent.Desc(notebook.FieldCreatedAt))
		}
	default:
		q = q.Order(ent.Desc(notebook.FieldUpdatedAt))
	}
	list, err := q.Offset(offset).Limit(limit).All(ctx)
	if err != nil {
		return nil, err
	}
	metas := make([]NotebookMeta, 0, len(list))
	for _, nb := range list {
		cellCount := 0
		if c, ok := nb.Content["cells"]; ok {
			if cells, ok := c.([]interface{}); ok {
				cellCount = len(cells)
			}
		}
		var desc *string
		if nb.Description != nil && *nb.Description != "" {
			desc = nb.Description
		}
		metas = append(metas, NotebookMeta{
			ID:          nb.ID.String(),
			Title:       nb.Title,
			Description: desc,
			IsPublic:    nb.IsPublic,
			CreatedAt:   nb.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   nb.UpdatedAt.Format(time.RFC3339),
			CellCount:   cellCount,
		})
	}
	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}
	return &ListNotebooksResult{
		Notebooks: metas,
		Pagination: Pagination{
			Page: page, Limit: limit, Total: total, TotalPages: totalPages,
		},
	}, nil
}

func (s *NotebookService) Get(ctx context.Context, notebookID string, userID uuid.UUID) (*ent.Notebook, error) {
	nbID, err := uuid.Parse(notebookID)
	if err != nil {
		return nil, err
	}
	nb, err := s.ent.Notebook.Query().Where(notebook.IDEQ(nbID)).WithOwner().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	if nb.OwnerID != userID {
		return nil, errors.New("access denied")
	}
	if nb.Edges.Owner == nil {
		owner, _ := s.ent.User.Get(ctx, nb.OwnerID)
		if owner != nil {
			nb.Edges.Owner = owner
		}
	}
	return nb, nil
}

func (s *NotebookService) Update(ctx context.Context, notebookID string, userID uuid.UUID, title, description *string, content map[string]interface{}, isPublic *bool) (*ent.Notebook, error) {
	nbID, err := uuid.Parse(notebookID)
	if err != nil {
		return nil, err
	}
	nb, err := s.ent.Notebook.Get(ctx, nbID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	if nb.OwnerID != userID {
		return nil, errors.New("access denied")
	}
	upd := s.ent.Notebook.UpdateOneID(nbID)
	if title != nil {
		t := *title
		if len(t) > 255 {
			t = t[:255]
		}
		upd = upd.SetTitle(t)
	}
	if description != nil {
		d := *description
		if len(d) > 1000 {
			d = d[:1000]
		}
		upd = upd.SetDescription(d)
	}
	if content != nil {
		upd = upd.SetContent(content)
	}
	if isPublic != nil {
		upd = upd.SetIsPublic(*isPublic)
	}
	updated, err := upd.Save(ctx)
	if err != nil {
		return nil, err
	}
	s.redis.Delete(ctx, notebookCachePrefix+notebookID)
	return updated, nil
}

func (s *NotebookService) Delete(ctx context.Context, notebookID string, userID uuid.UUID) error {
	nbID, err := uuid.Parse(notebookID)
	if err != nil {
		return err
	}
	nb, err := s.ent.Notebook.Get(ctx, nbID)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("not found")
		}
		return err
	}
	if nb.OwnerID != userID {
		return errors.New("access denied")
	}
	if err := s.ent.Notebook.DeleteOneID(nbID).Exec(ctx); err != nil {
		return err
	}
	s.redis.Delete(ctx, notebookCachePrefix+notebookID)
	s.redis.DeleteByPattern(ctx, "exec_session:"+notebookID)
	return nil
}
