package server

func (s *Server) StartWorkers() {
	go s.ThreadSaver()
	go s.SocketUpdater()
}

func (s *Server) ThreadSaver() {
	for {
		t := <-s.threadChannel
		s.store.SaveThread(t)
		s.sendChannel <- "thread"
	}
}

func (s *Server) SocketUpdater() {
	for {
		signal := <-s.sendChannel
		switch signal {
		case "thread":
			s.socketManager.Broadcast(s.socketManager.GetChatClients(), s.store.GetThreads())
		case "pair":
			s.socketManager.Broadcast(s.socketManager.GetPairClients(), s.text)
		}
	}
}
