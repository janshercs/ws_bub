package server

func (s *Server) StartWorkers() {
	go s.ThreadSaver()
	go s.SocketUpdater()
}

func (s *Server) ThreadSaver() {
	for {
		t := <-s.threadChannel
		s.store.SaveThread(t)
		s.sendChannel <- true
	}
}

func (s *Server) SocketUpdater() {
	for {
		signal := <-s.sendChannel
		if signal {
			s.socketManager.Broadcast(s.store.GetThreads())
		}
	}
}
