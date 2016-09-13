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

const (
	// SemVer is the current version.
	SemVer = "1.0.0"
)

// defaults
const (
	// DefaultWorkDir is the default working directory.
	DefaultWorkDir = "."

	// DefaultFileInputFilter is the default file input filter.
	DefaultFileInputFilter = ".jpg"

	// DefaultFileOutputPattern is the default output pattern for the file.`
	DefaultFileOutputPattern = "{DateTime.Year}{DateTime.Month}{DateTime.Day}_{Make}_{File.IndexByCaptureDate}.{File.Extension}"
)

// flags
var (
	flagWorkDir           = flag.String("workdir", DefaultWorkDir, "The working directory for operations.")
	flagInputFileFilter   = flag.String("filter", DefaultFileInputFilter, "The input file filter.")
	flagOutputFilePattern = flag.String("output", DefaultFileOutputPattern, "The file output pattern.")
	flagRecursive         = flag.Bool("recursive", false, "The filesystem visitor should recurse to sub directories.")
	flagDryRun            = flag.Bool("dryrun", true, "The print the output, do not rename/move the files.")
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

// --------------------------------------------------------------------------------
// Arguments
// --------------------------------------------------------------------------------

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

// --------------------------------------------------------------------------------
// Property Formatters
// --------------------------------------------------------------------------------

// TimestampProp returns a property of a timestamp.
func TimestampProp(timestamp time.Time, properties ...string) string {
	if len(properties) > 0 {
		switch properties[0] {
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
		case "Unix":
			return strconv.FormatInt(timestamp.Unix(), 10)
		case "Weekday":
			return fmt.Sprintf("%v", timestamp.Weekday())
		case "Offset":
			return timestamp.Location().String()
		}
	}
	return timestamp.Format(time.RFC3339)
}

// FileProp returns a file property.
func FileProp(fileMeta os.FileInfo, properties ...string) string {
	var value string
	if len(properties) > 0 {
		switch properties[0] {
		case "Name":
			{
				value = fileMeta.Name()
			}
		case "ModTime":
			{
				var subProperty string
				if len(properties) > 1 {
					subProperty = properties[1]
				}
				return TimestampProp(fileMeta.ModTime(), subProperty)
			}
		case "Size":
			{
				return strconv.FormatInt(fileMeta.Size(), 10)
			}
		}
	}
	return value
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

// GetExifData returns exif file meta for a given path.
func GetExifData(filePath string) (*exif.Exif, error) {
	fileContents, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer fileContents.Close()
	return exif.Decode(fileContents)
}

// ParseTagProperties returns the tag and relevant property.
func ParseTagProperties(outputTag string) (tag string, properties []string) {
	if strings.Contains(outputTag, ".") {
		parts := strings.Split(outputTag, ".")
		return parts[0], parts[1:]
	}
	return outputTag, nil
}

// ReplaceTagInPattern replaces a given tag in a given pattern.
func ReplaceTagInPattern(inputPattern, tag, value string) string {
	return strings.Replace(inputPattern, "{"+tag+"}", value, -1)
}

// GetFileTagValue gets a tag value from file metadata.
func GetFileTagValue(collector *DateIndexCollector, fileCaptureTime time.Time, filePath, tag string, properties ...string) (string, error) {
	var tagValue string
	fileMeta, err := os.Stat(filePath)
	if err != nil {
		return tagValue, err
	}

	if len(properties) > 0 {
		switch properties[0] {
		case "Index":
			{
				return fmt.Sprintf("%06d", collector.Len()), nil
			}
		case "IndexByCaptureYear":
			{
				fileIndex := collector.GetIndexByYear(fileCaptureTime)
				return fmt.Sprintf("%06d", fileIndex), nil
			}
		case "IndexByCaptureMonth":
			{
				fileIndex := collector.GetIndexByMonth(fileCaptureTime)
				return fmt.Sprintf("%06d", fileIndex), nil
			}
		case "IndexByCaptureDate":
			{
				fileIndex := collector.GetIndexByDay(fileCaptureTime)
				return fmt.Sprintf("%06d", fileIndex), nil
			}
		case "Extension":
			{
				return strings.Replace(filepath.Ext(fileMeta.Name()), ".", "", -1), nil
			}
		default:
			{
				return FileProp(fileMeta, properties...), nil
			}
		}
	}

	return tagValue, nil
}

// GetExifTagValue gets a tag value from exif metadata.
func GetExifTagValue(exifData *exif.Exif, tag string, properties ...string) (string, error) {
	var tagValue string
	exifTag, err := exifData.Get(exif.FieldName(tag))
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
		if len(properties) > 0 {
			tagValue = TimestampProp(timestamp, properties[0])
		}
	} else {
		tagValue = stringTagValue
	}

	return tagValue, nil
}

// GetTagValue returns the tag value for a given fileMeta.
func GetTagValue(indexCollector *DateIndexCollector, fileCaptureTime time.Time, exifData *exif.Exif, filePath, fileTag string) (string, error) {
	var tagValue string
	for _, outputTag := range strings.Split(fileTag, "|") {
		tag, properties := ParseTagProperties(outputTag)
		switch tag {
		case "File":
			fileTagValue, err := GetFileTagValue(indexCollector, fileCaptureTime, filePath, tag, properties...)
			if err != nil {
				continue
			}
			tagValue = fileTagValue
			break
		default:
			exifTagValue, err := GetExifTagValue(exifData, tag, properties...)
			if err != nil {
				continue
			}
			tagValue = exifTagValue
			break
		}
	}
	return tagValue, nil
}

// GetFileCaptureTime returns the capture time for a given image file.
func GetFileCaptureTime(filePath string) (time.Time, *exif.Exif, error) {
	var timestamp time.Time
	exifData, err := GetExifData(filePath)
	if err != nil {
		return timestamp, exifData, err
	}

	exifTag, err := exifData.Get(exif.DateTime)
	if err != nil {
		exifTag, err = exifData.Get(exif.DateTimeDigitized)
		if err != nil {
			exifTag, err = exifData.Get(exif.DateTimeOriginal)
		}
	}
	if err != nil {
		return timestamp, exifData, err
	}

	stringTagValue, err := exifTag.StringVal()
	if err != nil {
		return timestamp, exifData, err
	}
	timestamp, err = time.Parse(timestampFormat, stringTagValue)
	return timestamp, exifData, err
}

// IncrementCaptureIndex increments the capture index for a file based on
// its capture time.
func IncrementCaptureIndex(filePath string, collector *DateIndexCollector) (time.Time, *exif.Exif, error) {
	timestamp, exifData, err := GetFileCaptureTime(filePath)
	if err != nil {
		return timestamp, exifData, err
	}
	collector.Add(timestamp)
	return timestamp, exifData, nil
}

// ApplyPattern applies the rename pattern to the files.
func ApplyPattern(files, fileTags []string, outputFilePattern string) error {
	var collector = NewDateIndexCollector()
	for _, file := range files {
		fileCaptureTime, exifData, err := IncrementCaptureIndex(file, collector)

		outputFilename := outputFilePattern
		for _, tag := range fileTags {
			value, err := GetTagValue(collector, fileCaptureTime, exifData, file, tag)
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
