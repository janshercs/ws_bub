package server

type MemStore struct {
	threads []Thread
}

func (s *MemStore) SaveThread(thread Thread) {
	s.threads = append(s.threads, thread)
}

func (s *MemStore) GetThreads() []Thread {
	return s.threads
}
