package dkg

import (
	pb "mpc/gen/mpc"
	"sync"
)

type JobManager struct {
	mu          sync.RWMutex
	jobs        map[string]*Job
	idempotency map[string]string
}

func NewJobManager() *JobManager {
	return &JobManager{
		jobs:        make(map[string]*Job),
		idempotency: make(map[string]string),
	}
}

func (m *JobManager) GetOrCreateJob(idempotencyKey, jobID, walletID string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if existingJobID, exists := m.idempotency[idempotencyKey]; exists {
		return existingJobID, false
	}

	m.idempotency[idempotencyKey] = jobID
	m.jobs[jobID] = &Job{
		JobID:    jobID,
		WalletID: walletID,
		Status:   pb.GetKeyGenStatusResponse_IN_PROGRESS,
	}

	return jobID, true
}

func (m *JobManager) GetJob(jobID string) (*Job, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if job, ok := m.jobs[jobID]; ok {
		return job, true
	}

	return nil, false
}

func (m *JobManager) CompleteJob(jobID string, res *KeyGenResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if job, ok := m.jobs[jobID]; ok {
		job.Status = pb.GetKeyGenStatusResponse_COMPLETED
		job.PublicKeyHex = res.PublicKeyHex
		job.BitcoinAddress = res.BitcoinAddress
		job.XPub = res.XPub
		job.ChainCodeHex = res.ChainCodeHex
	}
}

func (m *JobManager) FailJob(jobID, errMsg string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if job, ok := m.jobs[jobID]; ok {
		job.Status = pb.GetKeyGenStatusResponse_FAILED
		job.ErrMsg = errMsg
	}
}
