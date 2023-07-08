# Audiobooker

![badge](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/cslamar/824d4b8e587def8656b0f5920e743467/raw/coverage.json)

Golang audiobook creation tool with support for tagging based on directory structure, transcoding concurrency, and more!

## Prerequisites

### Installation Requirements

* ffmpeg (developed against `5.1.x` and `4.1.x`)

### Build Requirements

* Go (developed using `1.20`)

## Operations

The `audiobooker` command has a number of sub-commands for various actions

* Documentation around command and subcommands can be found [here](docs/cli-usage)
* Usage examples can be found [here](docs/EXAMPLES.md) 

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

Audiobooker is able to parse the structure of the directories that hold the source audiobook files and embed that data in the final audiobook file.  This is done using a series of path pattern placeholders, more information can be found [here](docs/path-patterns.md).

### Other bits

* Automatic cover art will be applied if one of the following files are found in the media root: `cover.jpg`, `cover.png`, `folder.jpg`, `folder.png`
* Automatic description metadata will be applied if one of the following files are found in the media root: `description.txt` or `comment.txt` 


## Notes on macOS

I do not currently have an Apple developer account, therefore the macOS binary cannot be signed/notarized.  Until I decide to pay for one again you may do one of the following:

- Download the source and build it locally.  This will avoid the whole quarantine thing all together.
- Download the release zip file from the releases section using either `curl` or `wget`, this workaround _should_ avoid the quarantine issue that happens when downloading through a browser.

```shell
curl -OL https://github.com/cslamar/audiobooker/releases/download/VERSION_TAG/audiobooker_macos_universal.zip
```

- Download from the releases section of this repo, in a browser, and run the following command in the directory where you've extracted the `audiobooker` binary:

```shell
xattr -d com.apple.quarantine audiobooker
```

That will remove the quarantine attribute and things will work as expected.  

This is **not** optimal.  

You should **not** do this with most software.  

Unfortunately since Apple doesn't allow notarizing of open source software unless you pay for the $99/year privilege.  This is open source, you can see the code, you see the build pipeline, etc.  I'm planning on leaving the un-notarized binaries in the releases section for ease of use, as I'd love to get more peoples' input and exposure.

`/end_rant`


## Inspiration and Special Thanks

This project was heavily inspired by [m4b-tool](https://github.com/sandreas/m4b-tool), it was not a lack of quality in that project that made me want to write my own, but I thought it'd be fun and wanted to write it in Go personally.  Much credit goes there and where imitation is found, it's completely out of flattery to someone that has design a great project!

A special thanks goes to [Sorrow446](https://github.com/Sorrow446) for allowing me to use their implementation of an MP4 tagging solution as a basis for my own.
