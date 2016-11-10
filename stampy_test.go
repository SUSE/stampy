package stampy

import (
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStampCreate(t *testing.T) {
	t.Parallel()

	expected := `,origin,no-series,no-event`
	assert := assert.New(t)

	// Create the path for file which does not exist
	workDir, err := ioutil.TempDir("", "test_stamp_create")
	assert.Nil(err)
	defer os.RemoveAll(workDir) // clean up
	series := path.Join(workDir, "makethis")

	// Create timestamp
	err = Stamp(series, "origin", "no-series", "no-event")
	assert.Nil(err)

	// Check that the file now exists
	sFile, err := os.Open(series)
	if assert.Nil(err) {
		// And check its contents.
		contents, err := ioutil.ReadFile(series)
		assert.Nil(err)
		assert.Contains(string(contents), expected)
	}

	// Cleanup
	err = sFile.Close()
	assert.Nil(err)
}

func TestStampAppend(t *testing.T) {
	t.Parallel()

	expected := `,origin,no-series,no-event\n.*,source,no-series,the-event`
	assert := assert.New(t)

	// Create the path for file which does not exist
	workDir, err := ioutil.TempDir("", "test_stamp_append")
	assert.Nil(err)
	defer os.RemoveAll(workDir) // clean up
	series := path.Join(workDir, "makethis")

	// Create timestamp
	err = Stamp(series, "origin", "no-series", "no-event")
	assert.Nil(err)

	// Create another timestamp
	err = Stamp(series, "source", "no-series", "the-event")
	assert.Nil(err)

	// Check that the file now exists
	sFile, err := os.Open(series)
	if assert.Nil(err) {
		// And check its contents.
		contents, err := ioutil.ReadFile(series)
		assert.Nil(err)
		assert.Regexp(regexp.MustCompile(expected), string(contents))
	}

	// Cleanup
	err = sFile.Close()
	assert.Nil(err)
}

func TestStampCSVSpecials(t *testing.T) {
	t.Parallel()

	expected := `,",cram","""no-series","an,event"`
	assert := assert.New(t)

	workDir, err := ioutil.TempDir("", "test_stamp_csvspecial")
	assert.Nil(err)
	defer os.RemoveAll(workDir) // clean up
	series := path.Join(workDir, "makethis")

	// Create timestamp, use characters special to CSV format
	err = Stamp(series, `,cram`, `"no-series`, "an,event")
	assert.Nil(err)

	// Check that the file now exists
	sFile, err := os.Open(series)
	if assert.Nil(err) {
		// And check its contents.
		contents, err := ioutil.ReadFile(series)
		assert.Nil(err)
		assert.Contains(string(contents), expected)
	}

	// Cleanup
	err = sFile.Close()
	assert.Nil(err)
}
