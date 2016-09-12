package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"strings"

	"github.com/rwcarlsen/goexif/exif"
)

// defaults
const (
	// DefaultWorkDir is the default working directory.
	DefaultWorkDir = "."

	// DefaultFileInputFilter is the default file input filter.
	DefaultFileInputFilter = ".jpg"

	// DefaultFileOutputPattern is the default output pattern for the file.`
	DefaultFileOutputPattern = "{DateTime.Year}{DateTime.Month}{DateTime.Day}_{File.Index}.{File.Extension}"
)

// flags
var (
	flagWorkDir           = flag.String("workdir", DefaultWorkDir, "The working directory for operations.")
	flagInputFileFilter   = flag.String("filter", DefaultFileInputFilter, "The input file filter.")
	flagOutputFilePattern = flag.String("output", DefaultFileOutputPattern, "The file output pattern.")
	flagRecursive         = flag.Bool("recursive", false, "The filesystem visitor should recurse to sub directories.")
	flagDryRun            = flag.Bool("dryrun", false, "The print the output, do not rename/move the files.")
)

// fieldTypes
var (
	timestampFields = map[exif.FieldName]bool{
		exif.DateTime:          true,
		exif.DateTimeOriginal:  true,
		exif.DateTimeDigitized: true,
	}
)

const (
	timestampFormat = `2006:01:02 15:04:05`
)

// ArgsWorkDir returns the default working directory.
func ArgsWorkDir() string {
	if flagWorkDir == nil {
		return DefaultWorkDir
	}
	return *flagWorkDir
}

// ArgsWorkDirAbsolute is the absolute work directory.
func ArgsWorkDirAbsolute() (string, error) {
	return filepath.Abs(ArgsWorkDir())
}

// ArgsInputFileFilter is the input file filter.
func ArgsInputFileFilter() string {
	if flagInputFileFilter != nil {
		return *flagInputFileFilter
	}
	return DefaultFileInputFilter
}

// ArgsOutputFilePattern is the output file pattern.
func ArgsOutputFilePattern() string {
	if flagOutputFilePattern != nil {
		return *flagOutputFilePattern
	}
	return DefaultFileOutputPattern
}

// ArgsRecursive returns if the filesystem visitor should be recursive.
func ArgsRecursive() bool {
	if flagRecursive != nil {
		return *flagRecursive
	}
	return false
}

// ArgsDryRun returns if the files should be moved or just printed.
func ArgsDryRun() bool {
	if flagDryRun != nil {
		return *flagDryRun
	}
	return false
}

// TimestampProp returns a property of a timestamp.
func TimestampProp(timestamp time.Time, property string) string {
	switch property {
	case "Year":
		return strconv.Itoa(timestamp.Year())
	case "Month":
		return fmt.Sprintf("%02d", int(timestamp.Month()))
	case "Day":
		return fmt.Sprintf("%02d", timestamp.Day())
	case "Hour":
		return fmt.Sprintf("%02d", timestamp.Hour())
	case "Minute":
		return fmt.Sprintf("%02d", timestamp.Minute())
	case "Second":
		return fmt.Sprintf("%02d", timestamp.Second())
	case "Nanosecond":
		return strconv.Itoa(timestamp.Nanosecond())
	case "Offset":
		return timestamp.Location().String()
	}
	return timestamp.Format(time.RFC3339)
}

// FileProp returns a file property.
func FileProp(index int, fileMeta os.FileInfo, property string) string {
	switch property {
	case "Index":
		{
			return strconv.Itoa(index)
		}
	case "Name":
		{
			return fileMeta.Name()
		}
	case "ModTimeUnix":
		{
			return strconv.FormatInt(fileMeta.ModTime().Unix(), 10)
		}
	}
	return property
}

// ExtractFileOutputTags extracts the tags from a file pattern.
func ExtractFileOutputTags(filePattern string) []string {
	var tags []string
	state := 0
	var tag *bytes.Buffer
	for _, r := range filePattern {
		switch state {
		case 0:
			{
				if r == rune('{') {
					tag = bytes.NewBuffer([]byte{})
					state = 1
				}
			}
		case 1:
			{
				if r == rune('}') {
					tags = append(tags, tag.String())
					state = 0
				}
				tag.WriteRune(r)
			}
		}
	}
	return tags
}

// FilesInDirectoryWithFilter returns the files in a directory with a given filter.
func FilesInDirectoryWithFilter(directoryPath, fileFilter string) []string {
	var files []string

	fileFilterRegex := regexp.MustCompile(fileFilter)

	filepath.Walk(directoryPath, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}
		if fileFilterRegex.MatchString(path) {
			files = append(files, path)
		}
		return nil
	})

	return files
}

// GetFileTagValue gets a tag value from file metadata.
func GetFileTagValue(fileIndex int, filePath string, tag, property string) (string, error) {
	var tagValue string
	fileMeta, err := os.Stat(filePath)
	if err != nil {
		return tagValue, err
	}

	switch property {
	case "Index":
		{
			return strconv.Itoa(fileIndex), nil
		}
	case "Extension":
		{
			return strings.Replace(filepath.Ext(fileMeta.Name()), ".", "", -1), nil
		}
	case "Size":
		{
			return strconv.FormatInt(fileMeta.Size(), 10), nil
		}
	case "ModTime":
		{
			return TimestampProp(fileMeta.ModTime(), property), nil
		}
	case "Name":
		{
			return fileMeta.Name(), nil
		}
	}

	return tagValue, nil
}

// GetFileExifMeta returns exif file meta for a given path.
func GetFileExifMeta(filePath string) (*exif.Exif, error) {
	fileContents, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer fileContents.Close()
	return exif.Decode(fileContents)
}

// GetExifTagValue gets a tag value from exif metadata.
func GetExifTagValue(filePath string, tag, property string) (string, error) {
	var tagValue string
	fileMeta, err := GetFileExifMeta(filePath)
	if err != nil {
		return tagValue, err
	}
	exifTag, err := fileMeta.Get(exif.FieldName(tag))
	if err != nil {
		return tagValue, err
	}

	stringTagValue, err := exifTag.StringVal()
	if err != nil {
		return tagValue, err
	}

	if _, isTimestampField := timestampFields[exif.FieldName(tag)]; isTimestampField {
		timestamp, err := time.Parse(timestampFormat, stringTagValue)
		if err != nil {
			return tagValue, err
		}
		tagValue = TimestampProp(timestamp, property)
	} else {
		tagValue = stringTagValue
	}

	return tagValue, nil
}

// ParseTagProperties returns the tag and relevant property.
func ParseTagProperties(outputTag string) (tag, parameter string) {
	if strings.Contains(outputTag, ".") {
		parts := strings.Split(outputTag, ".")
		return parts[0], parts[1]
	}
	return outputTag, ""
}

// GetTagValue returns the tag value for a given fileMeta.
func GetTagValue(fileIndex int, filePath, fileTag string) (string, error) {
	var tagValue string
	for _, outputTag := range strings.Split(fileTag, "|") {
		tag, property := ParseTagProperties(outputTag)
		switch tag {
		case "File":
			fileTagValue, err := GetFileTagValue(fileIndex, filePath, tag, property)
			if err != nil {
				continue
			}
			tagValue = fileTagValue
			break
		default:
			exifTagValue, err := GetExifTagValue(filePath, tag, property)
			if err != nil {
				continue
			}
			tagValue = exifTagValue
			break
		}
	}
	return tagValue, nil
}

// ReplaceTagInPattern replaces a given tag in a given pattern.
func ReplaceTagInPattern(inputPattern, tag, value string) string {
	return strings.Replace(inputPattern, "{"+tag+"}", value, -1)
}

// ApplyPattern applies the rename pattern to the files.
func ApplyPattern(files, fileTags []string, outputFilePattern string) error {
	var err error
	for fileIndex, file := range files {
		outputFilename := outputFilePattern
		for _, tag := range fileTags {
			value, err := GetTagValue(fileIndex, file, tag)
			if err != nil {
				return err
			}
			outputFilename = ReplaceTagInPattern(outputFilename, tag, value)
		}

		if ArgsDryRun() {
			fmt.Printf("%s => %s\n", file, outputFilename)
		} else {
			err = os.Rename(file, outputFilename)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	flag.Parse()

	// - get all files in WorkDirAbsolute() that match the input filter
	workDir, err := ArgsWorkDirAbsolute()
	if err != nil {
		log.Fatal(err)
	}

	files := FilesInDirectoryWithFilter(workDir, ArgsInputFileFilter())
	fileTags := ExtractFileOutputTags(ArgsOutputFilePattern())

	err = ApplyPattern(files, fileTags, ArgsOutputFilePattern())
	if err != nil {
		log.Fatal(err)
	}
}
