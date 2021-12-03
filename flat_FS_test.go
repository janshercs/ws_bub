package server_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"server"
	"testing"
)

func TestFlatFileSystemStore(t *testing.T) {
	t.Run("file system initiation with threads", func(t *testing.T) {
		want := []server.Thread{
			{ID: 0, Content: "Hi", User: "Anna", UpVotesCount: 1, DownVotesCount: 1},
			{ID: 1, Content: "Bye", User: "Bob", UpVotesCount: 1, DownVotesCount: 1},
		}

		tmpfile, removeFile := createTempFile(t)
		defer removeFile()

		tmpfile.Write(ThreadsToBytes(t, want))

		store := getNewFFS(t, tmpfile)

		threads := store.GetThreads()
		if len(threads) == 0 {
			t.Fatal("no threads returned")
		}
		assertThreads(t, threads, want)
	})

	t.Run("file system able to save new threads", func(t *testing.T) {
		testThread := server.Thread{ID: 0, Content: "Hi", User: "Anna", UpVotesCount: 1, DownVotesCount: 1}
		want := []server.Thread{
			testThread,
		}

		tmpfile, removeFile := createTempFile(t)
		defer removeFile()

		store := getNewFFS(t, tmpfile)

		store.SaveThread(testThread)

		threads := store.GetThreads()
		if len(threads) == 0 {
			t.Fatal("no threads returned")
		}
		assertThreads(t, threads, want)
	})

	t.Run("file system able to add new threads to existing ones", func(t *testing.T) {
		testThread := server.Thread{ID: 3, Content: "Hi", User: "Banna", UpVotesCount: 1, DownVotesCount: 1}
		originalThreads := []server.Thread{
			{ID: 0, Content: "Hi", User: "Anna", UpVotesCount: 1, DownVotesCount: 1},
			{ID: 1, Content: "Bye", User: "Bob", UpVotesCount: 1, DownVotesCount: 1},
		}

		tmpfile, removeFile := createTempFile(t)
		defer removeFile()

		tmpfile.Write(ThreadsToBytes(t, originalThreads))

		store := getNewFFS(t, tmpfile)

		store.SaveThread(testThread)

		threads := store.GetThreads()
		if len(threads) == 0 {
			t.Fatal("no threads returned")
		}

		want := append(originalThreads, testThread)
		assertThreads(t, threads, want)
	})
}

func TestDatabaseWriter(t *testing.T) {
	file, clean := createTempFile(t)
	defer clean()
	file.Write([]byte("12345"))

	writer := server.NewFFSWriter(file)
	writer.Write([]byte("abc"))

	file.Seek(0, 0)
	newFileContents, _ := ioutil.ReadAll(file)

	got := string(newFileContents)
	want := "abc"

	if got != want {
		t.Errorf("want %q got %q", want, got)
	}

}

func createTempFile(t testing.TB) (*os.File, func()) {
	tmpfile, err := ioutil.TempFile("", "db")

	if err != nil {
		t.Fatalf("could not create temp file %v", err)
	}

	removeFile := func() {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
	}

	return tmpfile, removeFile
}

func ThreadsToBytes(t testing.TB, threads []server.Thread) []byte {
	t.Helper()

	var indexBuffer bytes.Buffer
	encoder := json.NewEncoder(&indexBuffer)

	if err := encoder.Encode(threads); err != nil {
		t.Errorf("Unable to convert thread to bytes, %v", err)
	}

	return indexBuffer.Bytes()
}

func getNewFFS(t testing.TB, file *os.File) *server.FlatFileSystem {
	ffs, err := server.NewFFS(file)

	if err != nil {
		t.Errorf("Unable to make new FFS, %v", err)
	}

	return ffs
}
