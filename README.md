# image-rename
`image-rename` is a customizable image renamer that uses image exif data to form the names.

## Installation 

Either download a pre-built binary for your platform or install it with `go get`

## Usage 

To use the utility, first test it out in a directory of your choosing: 

```
> image-rename --dryrun=true
```

## Output Format

You can optionally, but it is very recommended, provide a tokenized output format. 

The default output format is as follows:

```{DateTimeDigitized.Year}{DateTimeDigitized.Month}{DateTimeDigitized.Day}_{Make}_{File.IndexByCaptureDate}.{File.Extension}```

Notice a couple things; 1) we can specify individual components of a given date time field (in this example, `DateTimeDigitized`) with the `DateTimeDigitized.Year` property notation. 2) We can use a special "File" tag to access additional information outside what is provided by Exif. In the above case we're using the index as bucketed by capture date.

In addition to the standard exif fields provided by [goexif](http://github.com/rwcarlsen/goexif/exif) there are a couple custom ones you can use:

- `File.Index` : The index of the file in the directory.
- `File.IndexByCaptureYear` : The index of the file as bucketed by the capture year.
- `File.IndexByCaptureMonth` : The index of the file as bucketed by the capture year and month.
- `File.IndexByCaptureDate` : The index of the file as bucketed by the capture year, month, and day.
- `File.Extension` : The original file extension for the file.
- `File.Size` : The size in bytes of the file. 
- `File.ModTime.*` : The datetime field corresponding to the last modification time; you can use standard date time properties on this.
- `File.Name` : The original file name.


