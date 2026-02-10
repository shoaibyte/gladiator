package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gladiator/ent"
	"gladiator/ent/notebook"
	"gladiator/ent/notebookshare"
	"gladiator/ent/user"
)

type ShareService struct {
	ent *ent.Client
}

func NewShareService(entClient *ent.Client) *ShareService {
	return &ShareService{ent: entClient}
}

func (s *ShareService) Share(ctx context.Context, notebookID string, sharedByUserID uuid.UUID, email string, permission notebookshare.Permission) error {
	nb, err := s.ent.Notebook.Get(ctx, uuid.MustParse(notebookID))
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("not found")
		}
		return err
	}
	if nb.OwnerID != sharedByUserID {
		return errors.New("access denied")
	}
	u, err := s.ent.User.Query().Where(user.EmailEQ(email)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("user not found")
		}
		return err
	}
	if u.ID == sharedByUserID {
		return errors.New("cannot share with yourself")
	}
	_, err = s.ent.NotebookShare.Create().
		SetNotebookID(nb.ID).
		SetSharedWithUserID(u.ID).
		SetSharedByUserID(sharedByUserID).
		SetPermission(permission).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return errors.New("already shared")
		}
		return err
	}
	return nil
}

func (s *ShareService) ListShares(ctx context.Context, notebookID string, userID uuid.UUID) ([]*ent.NotebookShare, error) {
	nbID, _ := uuid.Parse(notebookID)
	nb, err := s.ent.Notebook.Get(ctx, nbID)
	if err != nil {
		return nil, err
	}
	if nb.OwnerID != userID {
		return nil, errors.New("access denied")
	}
	return s.ent.NotebookShare.Query().Where(notebookshare.NotebookIDEQ(nbID)).WithSharedWith().All(ctx)
}

func (s *ShareService) UpdatePermission(ctx context.Context, shareID string, notebookID string, userID uuid.UUID, permission notebookshare.Permission) error {
	nb, _ := s.ent.Notebook.Get(ctx, uuid.MustParse(notebookID))
	if nb.OwnerID != userID {
		return errors.New("access denied")
	}
	sID, _ := uuid.Parse(shareID)
	_, err := s.ent.NotebookShare.UpdateOneID(sID).SetPermission(permission).Save(ctx)
	return err
}

func (s *ShareService) Revoke(ctx context.Context, shareID string, notebookID string, userID uuid.UUID) error {
	nb, _ := s.ent.Notebook.Get(ctx, uuid.MustParse(notebookID))
	if nb.OwnerID != userID {
		return errors.New("access denied")
	}
	sID, _ := uuid.Parse(shareID)
	return s.ent.NotebookShare.DeleteOneID(sID).Exec(ctx)
}

func (s *NotebookService) ListPublic(ctx context.Context, page, limit int, search string) (*ListNotebooksResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	q := s.ent.Notebook.Query().Where(notebook.IsPublicEQ(true))
	if search != "" {
		q = q.Where(notebook.Or(
			notebook.TitleContainsFold(search),
			notebook.DescriptionContainsFold(search),
		))
	}
	total, _ := q.Clone().Count(ctx)
	list, _ := q.Order(ent.Desc(notebook.FieldUpdatedAt)).Offset(offset).Limit(limit).All(ctx)
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
			ID: nb.ID.String(), Title: nb.Title, Description: desc, IsPublic: true,
			CreatedAt: nb.CreatedAt.Format(time.RFC3339),
			UpdatedAt: nb.UpdatedAt.Format(time.RFC3339),
			CellCount:  cellCount,
		})
	}
	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}
	return &ListNotebooksResult{Notebooks: metas, Pagination: Pagination{Page: page, Limit: limit, Total: total, TotalPages: totalPages}}, nil
}

func (s *NotebookService) Fork(ctx context.Context, notebookID string, userID uuid.UUID) (*ent.Notebook, error) {
	nb, err := s.ent.Notebook.Query().Where(notebook.IDEQ(uuid.MustParse(notebookID))).Only(ctx)
	if err != nil {
		return nil, err
	}
	if !nb.IsPublic {
		return nil, errors.New("not found")
	}
	content := nb.Content
	if content == nil {
		content = defaultCells()
	}
	create := s.ent.Notebook.Create().
		SetOwnerID(userID).
		SetTitle(nb.Title + " (fork)").
		SetContent(content).
		SetIsPublic(false)
	if nb.Description != nil {
		create = create.SetDescription(*nb.Description)
	}
	return create.Save(ctx)
}
