package toolkit

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func TestTools_RandomString(t *testing.T) {
	var testTools Tools

	s := testTools.RandomString(10)
	if len(s) != 10 {
		t.Error("wrong length of random string returned")
	}
}

var uploadTests = []struct {
	name          string
	allowedTypes  []string
	renameFile    bool
	errorExpected bool
}{
	{
		name:          "allowed no rename",
		allowedTypes:  []string{"image/jpeg", "image/png"},
		renameFile:    false,
		errorExpected: false,
	},
	{
		name:          "allowed rename",
		allowedTypes:  []string{"image/jpeg", "image/png"},
		renameFile:    true,
		errorExpected: false,
	},
	{
		name:          "not allowed",
		allowedTypes:  []string{"image/jpeg"},
		renameFile:    false,
		errorExpected: true,
	},
}

func TestTools_UploadFiles(t *testing.T) {
	for _, e := range uploadTests {
		// set up a pipe to avioid buttering
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)
		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer writer.Close()
			defer wg.Done()

			/// create the form data filed 'file'
			part, err := writer.CreateFormFile("file", "./testdata/img.png")
			if err != nil {
				t.Errorf("%s", err.Error())
			}

			f, err := os.Open("./testdata/img.png")
			if err != nil {
				t.Errorf("%s", err.Error())
			}
			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				t.Errorf("error decoding image %s", err)
			}

			err = png.Encode(part, img)
			if err != nil {
				t.Errorf("error encoding image %s", err)
			}
		}()

		// read from the pipe wich receives data
		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools
		testTools.AllowedFileTypes = e.allowedTypes

		uploadedFiles, err := testTools.UploadFiles(request, "./testdata/uploads/", e.renameFile)
		if err != nil && !e.errorExpected {
			t.Errorf("error uploading file: %v", err)
		}

		if !e.errorExpected {
			if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName)); os.IsNotExist(err) {
				t.Errorf("%s file not found: %s", e.name, err.Error())
			}

			// clean up the uploaded file
			_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName))
		}

		if !e.errorExpected && err != nil {
			t.Errorf("%s: error expected  but none received", e.name)
		}

		wg.Wait()
	}
}

func TestTools_UploadOneFile(t *testing.T) {
	// set up a pipe to avioid buttering
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)
	go func() {
		defer writer.Close()

		/// create the form data filed 'file'
		part, err := writer.CreateFormFile("file", "./testdata/img.png")
		if err != nil {
			t.Errorf("%s", err.Error())
		}

		f, err := os.Open("./testdata/img.png")
		if err != nil {
			t.Errorf("%s", err.Error())
		}
		defer f.Close()

		img, _, err := image.Decode(f)
		if err != nil {
			t.Errorf("error decoding image %s", err)
		}

		err = png.Encode(part, img)
		if err != nil {
			t.Errorf("error encoding image %s", err)
		}
	}()

	// read from the pipe wich receives data
	request := httptest.NewRequest("POST", "/", pr)
	request.Header.Add("Content-Type", writer.FormDataContentType())

	var testTools Tools

	uploadedFile, err := testTools.UploadOneFile(request, "./testdata/uploads/", true)
	if err != nil {
		t.Error(err)
	}

	if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFile.NewFileName)); os.IsNotExist(err) {
		t.Errorf("file not found: %s", err.Error())
	}

	// clean up the uploaded file
	_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFile.NewFileName))
}

func TestTools_CreateDirIfNotExist(t *testing.T) {
	dir := "./testdata/myDir"
	var testTools Tools

	err := testTools.CreateDirIfNotExist(dir)
	if err != nil {
		t.Errorf("error creating directory: %s", err.Error())
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("directory not found: %s", err.Error())
	}

	// clean up the created directory
	_ = os.RemoveAll(dir)
}

var slugifyTests = []struct {
	name          string
	s             string
	expected      string
	errorExpected bool
}{
	{name: "valid slug", s: "Hello World!", expected: "hello-world", errorExpected: false},
	{name: "empty slug", s: "", expected: "", errorExpected: true},
	{name: "complex string", s: "Now is the time for all GOOD men! + fish & such &^123", expected: "now-is-the-time-for-all-good-men-fish-such-123", errorExpected: false},
	{name: "japanese string", s: "こんにちは世界", expected: "", errorExpected: true},
	{name: "japanese string and roman characters", s: "Hello World こんにちは世界", expected: "hello-world", errorExpected: false},
}

func TestTools_Slugify(t *testing.T) {
	var testTools Tools

	for _, e := range slugifyTests {
		slugified, err := testTools.Slugify(e.s)
		if err != nil && !e.errorExpected {
			t.Errorf("%s error received when none expected: %s", e.name, err.Error())
		}

		if !e.errorExpected && slugified != e.expected {
			t.Errorf("%s: expected %s but got %s", e.name, e.expected, slugified)
		}

		if e.errorExpected && err == nil {
			t.Errorf("%s: error expected but none received", e.name)
		}
	}
}

func TestTools_DownloadStaticFile(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	var testTool Tools
	testTool.DownloadStaticFile(rr, req, "./testdata", "pic.jpg", "pappy.jpg")

	res := rr.Result()
	defer res.Body.Close()

	if res.Header["Content-Length"][0] != "98827" {
		t.Errorf("Content-Length header is not correct, got %s", res.Header["Content-Length"][0])
	}

	if res.Header["Content-Disposition"][0] != "attachment; filename=\"pappy.jpg\"" {
		t.Errorf("Content-Disposition header is not correct, got %s", res.Header["Content-Disposition"][0])
	}

	_, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("error reading response body: %s", err.Error())
	}
}
