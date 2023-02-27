# Audiobooker

![badge](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/cslamar/824d4b8e587def8656b0f5920e743467/raw/coverage.json)

Golang audiobook creation tool with support for tagging based on directory structure, transcoding concurrency, and more!

## Prerequisites

### Installation Requirements

* ffmpeg (developed against `5.1.x` and `4.1.x`)

### Build Requirements

* Go (developed using `1.19`)

## Operations

The `audiobooker` command has a number of sub-commands for various actions

* Documentation around command and subcommands can be found [here](docs)
* Usage examples can be found [here](EXAMPLES.md) 

### Environment Variables for Commands

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

**Note:** `path-pattern` in `batch` works different from `bind`.  In `batch` the pattern starts **from** the path specified in `source-files-root`

### Other bits

Automatic cover art will be applied if one of the following files are found in the media root: `cover.jpg`, `cover.png`, `folder.jpg`, `folder.png`

## Inspiration and Special Thanks

This project was heavily inspired by [m4b-tool](https://github.com/sandreas/m4b-tool), it was not a lack of quality in that project that made me want to write my own, but I thought it'd be fun and wanted to write it in Go personally.  Much credit goes there and where imitation is found, it's completely out of flattery to someone that has design a great project!

A special thanks goes to [Sorrow446](https://github.com/Sorrow446) for allowing me to use their implementation of an MP4 tagging solution as a basis for my own.
