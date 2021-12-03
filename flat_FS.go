package server

import (
	"encoding/json"
	"fmt"
	"os"
)

func NewFFSFromPath(path string) (*FlatFileSystem, func(), error) {
	db, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)

	if err != nil {
		return nil, nil, fmt.Errorf("Error opening %s file, %v", path, err)
	}

	ffs, err := NewFFS(db)
	if err != nil {
		return nil, nil, fmt.Errorf("Error initiating Flat File System from file, %s %v", path, err)
	}

	return ffs, func() { db.Close() }, nil
}

func NewFFS(file *os.File) (*FlatFileSystem, error) {
	file.Seek(0, 0)
	err := initialiseFlatFileDB(file)
	if err != nil {
		return nil, fmt.Errorf("Unable to initialize file for FFS, %v", err)
	}

	threads, err := GetThreadsFromReader(file)
	if err != nil {
		return nil, fmt.Errorf("Unable to get threads from input, %v", err)
	}

	return &FlatFileSystem{database: json.NewEncoder(&FFSWriter{file: file}), threads: threads}, nil
}

func initialiseFlatFileDB(file *os.File) error {
	file.Seek(0, 0)

	info, err := file.Stat()

	if err != nil {
		return fmt.Errorf("problem getting file info from file %s, %v", file.Name(), err)
	}

	if info.Size() == 0 {
		file.Write([]byte("[]"))
		file.Seek(0, 0)
	}
	return nil
}

type FlatFileSystem struct {
	database *json.Encoder
	threads  []Thread
}

func (f *FlatFileSystem) GetThreads() []Thread {
	return f.threads
}

func (f *FlatFileSystem) SaveThread(t Thread) {
	f.threads = append(f.threads, t)
	f.database.Encode(f.threads)
}

type FFSWriter struct {
	file *os.File
}

func NewFFSWriter(f *os.File) *FFSWriter {
	return &FFSWriter{file: f}
}

func (w *FFSWriter) Write(p []byte) (n int, err error) {
	w.file.Truncate(0)
	w.file.Seek(0, 0)
	return w.file.Write(p)
}
