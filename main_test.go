package main

import (
	"fmt"
	"testing"

	assert "github.com/blendlabs/go-assert"
)

func TestExtractFileTags(t *testing.T) {
	assert := assert.New(t)

	tags := ExtractFileOutputTags(DefaultFileOutputPattern)
	assert.Len(tags, 6, fmt.Sprintf("%#v", tags))
	assert.Equal("DateTime.Year", tags[0])
	assert.Equal("DateTime.Month", tags[1])
	assert.Equal("DateTime.Day", tags[2])
	assert.Equal("Make", tags[3])
	assert.Equal("File.IndexByCaptureDate", tags[4])
	assert.Equal("File.Extension", tags[5])
}

func TestReplaceTagInPattern(t *testing.T) {
	assert := assert.New(t)

	pattern := "{foo}_{bar}_{foo}"
	replaced := ReplaceTagInPattern(pattern, "foo", "123")
	assert.Equal("123_{bar}_123", replaced)
	assert.Equal("123_321_123", ReplaceTagInPattern(replaced, "bar", "321"))
}
