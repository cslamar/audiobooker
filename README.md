# Audiobooker

Golang audiobook creation tool with support for tagging based on directory structure, transcoding concurrency, and more!

## Installation

**Requirements:**

* ffmpeg

## Operations

The `audiobooker` command has a number of sub-commands for various actions

### Bind
Combine multiple audio files into an M4B audiobook file

```text
  files          Bind audiobook using each file as a chapter
  from-tags      Bind audiobook combining title tag of each file as chapter names
  split-chapters Splits a single audio file into chapters using a fixed length
```

### Batch
Perform batched operations on a pattern of directories for multiple audiobook binding

```text
  files          Bind audiobook using each file as a chapter
  from-tags      Bind audiobook combining title tag of each file as chapter names
  split-chapters Splits a single audio file into chapters using a fixed length
```

### Configurations

#### Global Bind Command Flags:

```text
  -f, --file-pattern string         The output filename, can be a combination of literal values and patterns
  -h, --help                        help for bind
  -j, --jobs int                    The number of concurrent transcoding process to run for conversion (don't exceed your cpu count) (default 1)
  -o, --output-directory string     The output directory for the final directory, can be combination of absolute values and path patterns
  -p, --path-pattern string         The pattern for metadata picked up via paths
      --scratch-files-path string   The location to generate the scratch directory
  -s, --source-files-path string    The path to directory of source files (must match path-pattern for metadata to work)
      --verbose-transcode           Enable output of all ffmpeg commands/operations
```

#### Global Batch Command Flags:

```text
  -f, --file-pattern string         The output filename, can be a combination of literal values and patterns
  -h, --help                        help for batch
  -j, --jobs int                    The number of concurrent transcoding process to run for conversion (don't exceed your cpu count) (default 1)
  -o, --output-directory string     The output directory for the final directory, can be combination of absolute values and path patterns
  -p, --path-pattern string         The pattern for metadata picked up via paths (starts from base of source-files-root)
      --scratch-files-path string   The location to generate the scratch directory
  -s, --source-files-root string    The path to directory of source files (must match path-pattern for metadata to work)
      --verbose-transcode
```

**Note:** `path-pattern` in `batch` works different from `bind`.  In `batch` the pattern starts **from** the path specified in `source-files-root`

#### Environment Variables for Commands

Configs can be set via environment variables.  The following are the currently supported variables for configuration:

| Variable              | Description                                                              |
|-----------------------|--------------------------------------------------------------------------|
| `JOBS`                | Number of concurrent transcode jobs to run                               |
| `OUTPUT_FILE_DEST`    | Directory path for output file                                           |
| `OUTPUT_FILE_PATTERN` | The output filename, can be a combination of literal values and patterns |
| `OUTPUT_PATH_PATTERN` | The path pattern template for dynamically created output directories     |
| `PATH_PATTERN`        | Input path pattern for generating tags from directory structure          |
| `SCRATCH_FILES_PATH`  | Directory path for temporary files                                       |


### Paths and Tagging

Tags are added either via the `--path-pattern` flag or `PATH_PATTERN` environment variable.  Supported path patterns are:

```text
AudioFile   = "%f"
Author      = "%a"
Genre       = "%g"
Narrator    = "%n"
ReleaseDate = "%y"
Series      = "%s"
SeriesPart  = "%p"
Title       = "%t"
```

The `Author` and `Title` are the only two that are currently required.

When structuring a `path-pattern` you must supply any hardcoded paths that aren't matched to a metadata parsing.

**Example Path**

```text
./media-src/files/Carl von Clausewitz/On War/1/Volume 1
```

**Example `path-pattern` string**

```text
./media-src/files/%a/%s/%p/%t
```

### Other bits

Automatic cover art will be applied if one of the following files are found in the media root: `cover.jpg`, `cover.png`, `folder.jpg`, `folder.png`

## Examples

#### Split Into Fixed Length Chapters (with patterned output directory)

```shell
audiobooker bind split-chapters \
  --chapter-length 5 \
  --path-pattern "./media-src/%a/%s/%p/%t" \
  --output-directory "./ab/final/%a/%s/%p" \
  --source-files-path "./media-src/Carl von Clausewitz/On War/1/Volume 1"
```

#### Create Audiobook From Structured Layout Compiling Chapters From Media Tags

```shell
audiobooker bind from-tags \
  --path-pattern "./media-src/%y/%a/%t" \
  --output-directory "./ab/final" \
  --source-files-path "./media-src/1903/Carl von Clausewitz/On War"
```

#### Create Audiobook From Structured Layout creating one chapter per input file

```shell
audiobooker bind files \
  --path-pattern "./media-src/%y/%a/%t" \
  --output-directory "./ab/final" \
  --source-files-path "./media-src/1903/Carl von Clausewitz/On War"
```

#### Create Audiobook From Structured Layout creating one chapter per input file with custom output path and filename

```shell
audiobooker bind files \
  --source-files-path "./media-src/files/chapter-by-file/1903/Carl von Clausewitz/On War" \
  --path-pattern "./media-src/files/chapter-by-file/%y/%a/%t" \
  --output-directory "./output/%a/%y/%t" \
  --file-pattern "%a - %t"
```

Would create the following directories and filename: `./output/Carl von Clausewitz/1903/On War/Carl von Clausewitz - On War.m4b`

#### Batch a Collection of Books at Once, with One Chapter per File

```shell
audiobooker batch files \
  --source-files-root "test-data/files/batching" \
  --path-pattern "%a/%s/%p/%t" \
  --output-directory "./ab/output/%a/%s/%p" \
  --file-pattern="%t" \
  --title-tag
```

**Where**
* `source-files-root` - path to top most directory containing source audio files
* `path-pattern` - metadata paths to map for each book
* `output-directory` - combination of static paths and metadata pattern paths, created dynamically, to output final books
* `file-pattern` - output name for final audiobook file
* `title-tag` - use the source audio file's metadata `title` tag as the chapter name in the new audiobook file

# Inspiration and Thanks

This project was heavily inspired by [m4b-tool](https://github.com/sandreas/m4b-tool), it was not a lack of quality in that project that made me want to write my own, but I thought it'd be fun and wanted to write it in Go personally.  Much credit goes there and where imitation is found, it's completely out of flattery to someone that has design a great project!
