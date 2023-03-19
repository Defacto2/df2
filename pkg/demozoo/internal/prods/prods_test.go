package prods_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/demozoo/internal/prods"
	"github.com/stretchr/testify/assert"
)

var (
	ErrAdd = errors.New("invalid add argument")
)

const (
	cd      = "Content-Disposition"
	modDate = "Wed, 30 Apr 2012 16:29:51 -0500"
)

func mocker(add string) (http.Header, error) {
	// source: https://blog.questionable.services/article/testing-http-handlers-go/
	var header http.Header
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/mock-header", nil)
	defer cancel()
	if err != nil {
		return header, err
	}
	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	var handler http.HandlerFunc
	switch add {
	case "cd":
		handler = http.HandlerFunc(mockContentDisposition)
	case "fn":
		handler = http.HandlerFunc(mockFilename)
	case "fn1":
		handler = http.HandlerFunc(mockFilename1)
	case "il":
		handler = http.HandlerFunc(mockInline)
	default:
		return header, fmt.Errorf("mock header %q: %w", add, ErrAdd)
	}
	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)
	return rr.Header(), err
}

func mockContentDisposition(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Add(cd, "attachment")
}

func mockFilename1(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Add(cd, "attachment; filename*=example.zip;")
	w.Header().Add("modification-date", modDate)
}

func mockFilename(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Add(cd, "attachment; filename=example.zip;")
	w.Header().Add("modification-date", modDate)
}

func mockInline(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Add(cd, "inline")
}

func Test_filename(t *testing.T) {
	s := prods.Filename(nil)
	assert.Equal(t, "", s)

	cd, err := mocker("cd")
	assert.Nil(t, err)
	s = prods.Filename(cd)
	assert.Equal(t, "", s)

	fn, err := mocker("fn")
	assert.Nil(t, err)
	s = prods.Filename(fn)
	assert.Equal(t, "example.zip", s)

	fn1, err := mocker("fn1")
	assert.Nil(t, err)
	s = prods.Filename(fn1)
	assert.Equal(t, "example.zip", s)

	il, err := mocker("il")
	assert.Nil(t, err)
	s = prods.Filename(il)
	assert.Equal(t, "", s)
}

func TestProductionsAPIv1_DownloadLink(t *testing.T) {
	p := prods.ProductionsAPIv1{}
	n, l := p.DownloadLink(nil)
	assert.Equal(t, "", n)
	assert.Equal(t, "", l)

	p = example1
	n, l = p.DownloadLink(io.Discard)
	assert.Equal(t, "feestje.zip", n)
	assert.Contains(t, l, "/parties/2000/ambience00/demo/feestje.zip")

	p = example2
	n, l = p.DownloadLink(io.Discard)
	assert.Equal(t, "the_untouchables_bbs7.zip", n)
	assert.Contains(t, l, "/demos/compilations/lost_found_and_more/bbs/the_untouchables_bbs7.zip")
}

func TestProductionsAPIv1_PouetID(t *testing.T) {
	p := prods.ProductionsAPIv1{}
	i, c, err := p.PouetID(false)
	assert.Nil(t, err)
	assert.Equal(t, 0, i)
	assert.Equal(t, 0, c)

	p = example3
	i, _, err = p.PouetID(false)
	assert.Nil(t, err)
	assert.Equal(t, 0, i)

	p = example1
	i, _, err = p.PouetID(false)
	assert.Nil(t, err)
	assert.Equal(t, 7084, i)
	assert.Equal(t, 0, c)
}

func TestProductionsAPIv1_Print(t *testing.T) {
	p := prods.ProductionsAPIv1{}
	err := p.Print(nil)
	assert.Nil(t, err)

	p = example1
	err = p.Print(io.Discard)
	assert.Nil(t, err)

	w := strings.Builder{}
	err = p.Print(&w)
	assert.Nil(t, err)
	assert.Contains(t, w.String(), "https://demozoo.org/api/v1/productions/1/?format=json")
}

func TestMutate(t *testing.T) {
	u, err := url.Parse("http://example.com")
	assert.Nil(t, err)
	u, err = prods.Mutate(u)
	assert.Nil(t, err)
	assert.Equal(t, "http://example.com", u.String())

	u, err = url.Parse("not-a-valid-url")
	assert.Nil(t, err)
	u, err = prods.Mutate(u)
	assert.Nil(t, err)
	assert.Equal(t, "not-a-valid-url", u.String())

	u, err = url.Parse("https://files.scene.org/view/someplace")
	assert.Nil(t, err)
	u, err = prods.Mutate(u)
	assert.Nil(t, err)
	assert.Equal(t, "https://files.scene.org/get:nl-http/someplace", u.String())
}

func TestParse(t *testing.T) {
	i, err := prods.Parse("")
	assert.NotNil(t, err)
	assert.Equal(t, 0, i)
	i, err = prods.Parse("https://www.pouet.net/prod.php?which=-1")
	assert.NotNil(t, err)
	assert.Equal(t, 0, i)
	i, err = prods.Parse("https://www.pouet.net/prod.php?which=abc")
	assert.NotNil(t, err)
	assert.Equal(t, 0, i)
	i, err = prods.Parse("https://www.pouet.net")
	assert.NotNil(t, err)
	assert.Equal(t, 0, i)
	i, err = prods.Parse("https://www.pouet.net/prod.php?which=30352")
	assert.Nil(t, err)
	assert.Equal(t, 30352, i)
}

func TestRandomName(t *testing.T) {
	r, err := prods.RandomName()
	assert.Nil(t, err)
	assert.NotEqual(t, "", r)
}

func Test_SaveName(t *testing.T) {
	s, err := prods.SaveName("")
	assert.Nil(t, err)
	assert.NotEqual(t, "", s)
	s, err = prods.SaveName("blob.txt")
	assert.Nil(t, err)
	assert.Equal(t, "blob.txt", s)
	s, err = prods.SaveName("https://example.com/path/to/some/file/down/here/blob.txt")
	assert.Nil(t, err)
	assert.Equal(t, "blob.txt", s)
}
