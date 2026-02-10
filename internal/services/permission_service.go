package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gladiator/ent"
	"gladiator/ent/notebookshare"
)

type Access struct {
	CanView   bool
	CanEdit   bool
	CanDelete bool
	IsOwner   bool
}

func (s *NotebookService) CheckNotebookAccess(ctx context.Context, notebookID string, userID uuid.UUID) (*Access, error) {
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
	if nb.OwnerID == userID {
		return &Access{CanView: true, CanEdit: true, CanDelete: true, IsOwner: true}, nil
	}
	share, err := s.ent.NotebookShare.Query().
		Where(notebookshare.NotebookIDEQ(nbID), notebookshare.SharedWithUserIDEQ(userID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("access denied")
		}
		return nil, err
	}
	canEdit := share.Permission == notebookshare.PermissionEdit || share.Permission == notebookshare.PermissionAdmin
	return &Access{
		CanView:   true,
		CanEdit:   canEdit,
		CanDelete: share.Permission == notebookshare.PermissionAdmin,
		IsOwner:   false,
	}, nil
}
