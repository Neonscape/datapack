package workspace

import (
	"fmt"
	"sync/atomic"
)

var epochCounter atomic.Uint64

type OperationState struct {
	ID        uint64 `json:"id"`
	Kind      string `json:"kind"`
	Phase     string `json:"phase"`
	Completed int    `json:"completed"`
	Total     int    `json:"total"`
	Running   bool   `json:"running"`
	Error     string `json:"error,omitempty"`
}

func (s *Service) GetOperationState() OperationState {
	s.operationMu.RLock()
	defer s.operationMu.RUnlock()
	return s.operation
}
func (s *Service) operationBusy() bool { return s.GetOperationState().Running }
func (s *Service) beginOperation(kind string) (func(error), error) {
	s.operationMu.Lock()
	if s.operation.Running {
		s.operationMu.Unlock()
		return nil, fmt.Errorf("another project operation is running")
	}
	id := s.operation.ID + 1
	s.operation = OperationState{ID: id, Kind: kind, Phase: "inputs", Running: true}
	s.operationMu.Unlock()
	return func(err error) {
		s.operationMu.Lock()
		defer s.operationMu.Unlock()
		if s.operation.ID != id {
			return
		}
		s.operation.Running = false
		if err != nil {
			s.operation.Error = err.Error()
			s.operation.Phase = "failed"
		} else {
			s.operation.Phase = "done"
			s.operation.Completed = s.operation.Total
		}
	}, nil
}
func (s *Service) operationProgress(phase string, completed, total int) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if s.operation.Running {
		s.operation.Phase = phase
		s.operation.Completed = completed
		s.operation.Total = total
	}
}
func (s *Service) markSourceUncertain() {
	s.sourceUncertain = true
	s.literalTables = nil
	s.successfulCompile = nil
	s.literalInputFingerprint = ""
	s.markEdited()
	view := *s.current
	view.SourceUncertain = true
	s.publishView(&view)
}
