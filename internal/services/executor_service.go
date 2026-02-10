package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gladiator/ent"
	"gladiator/ent/notebook"
	"gladiator/internal/database"
	"gladiator/pkg/executor"
)

type ExecutorService struct {
	entClient *ent.Client
	redis     *database.RedisClient
	exec      *executor.Executor
	session   *executor.SessionManager
}

func NewExecutorService(entClient *ent.Client, redis *database.RedisClient) *ExecutorService {
	exec := executor.NewExecutor()
	session := executor.NewSessionManager(
		redis.Get,
		func(ctx context.Context, key string, value string, ttl time.Duration) error {
			return redis.SetWithExpiry(ctx, key, value, ttl)
		},
		redis.Delete,
	)
	return &ExecutorService{
		entClient: entClient,
		redis:     redis,
		exec:      exec,
		session:   session,
	}
}

func (s *ExecutorService) ExecuteCell(ctx context.Context, notebookID string, userID string, cellID string, code string) (*executor.ExecutionResult, error) {
	nbID, err := uuid.Parse(notebookID)
	if err != nil {
		return nil, err
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	nb, err := s.entClient.Notebook.Query().Where(notebook.IDEQ(nbID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("notebook not found")
		}
		return nil, err
	}
	if nb.OwnerID != uid {
		return nil, errors.New("access denied")
	}
	current, _ := s.session.GetOrCreateSession(ctx, notebookID)
	var fullCode string
	if current == "" {
		fullCode = code
	} else {
		fullCode = current + "\n\n" + code
	}
	result, err := s.exec.Execute(ctx, fullCode)
	if err != nil {
		return nil, err
	}
	if result.Status == "ok" {
		_ = s.session.SetSession(ctx, notebookID, fullCode)
		// Update cell output in notebook content
		content := nb.Content
		if content == nil {
			content = map[string]interface{}{"cells": []interface{}{}}
		}
		cells, _ := content["cells"].([]interface{})
		for i, c := range cells {
			cm, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			if id, _ := cm["id"].(string); id == cellID {
				cm["output"] = result.Stdout
				if result.Stderr != "" {
					cm["output"] = result.Stdout + "\n" + result.Stderr
				}
				cm["executed_at"] = time.Now().Format(time.RFC3339)
				cells[i] = cm
				break
			}
		}
		content["cells"] = cells
		_, _ = s.entClient.Notebook.UpdateOneID(nbID).SetContent(content).SetLastExecutedAt(time.Now()).AddExecutionCount(1).Save(ctx)
	}
	return result, nil
}

func (s *ExecutorService) GetSessionInfo(ctx context.Context, notebookID string, userID string) (string, error) {
	if err := s.checkAccess(ctx, notebookID, userID); err != nil {
		return "", err
	}
	return s.session.SessionInfo(ctx, notebookID)
}

func (s *ExecutorService) ClearSession(ctx context.Context, notebookID string, userID string) error {
	if err := s.checkAccess(ctx, notebookID, userID); err != nil {
		return err
	}
	return s.session.ClearSession(ctx, notebookID)
}

func (s *ExecutorService) checkAccess(ctx context.Context, notebookID string, userID string) error {
	nbID, err := uuid.Parse(notebookID)
	if err != nil {
		return err
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	nb, err := s.entClient.Notebook.Get(ctx, nbID)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("not found")
		}
		return err
	}
	if nb.OwnerID != uid {
		return errors.New("access denied")
	}
	return nil
}

