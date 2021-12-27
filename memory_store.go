package server

type MemStore struct {
	threads Threads
}

func (s *MemStore) SaveThread(thread Thread) {
	s.threads = append(s.threads, thread)
}

func (s *MemStore) GetThreads() Threads {
	return s.threads
}
