## audiobooker batch split-chapters

Splits a single audio file into chapters using a fixed length

### Synopsis

Bind split-chapters will split a single file into a chapter marked audiobook file based on two options.  

First a static number (in minutes) can be passed in to make hard chapter marks at the specified duration.  Each mark will result in chapter metadata being created at those increments with the name "Chapter X" (where X in the index).

The other way that split-chapters can be used is if the existing file already has metadata embedded.  Passing in the '--use-embedded' flag will use that metadata when creating the chapters for the new audiobook file.

```
audiobooker batch split-chapters [flags]
```

### Options

```
  -c, --chapter-length int   chapter length in minutes (default 5)
      --generate-chapters    generate chapters and embed them in and existing .m4b audiobook (no transcoding required)
  -h, --help                 help for split-chapters
      --use-embedded         use existing embedded chapters
```

### Options inherited from parent commands

```
      --alert                       enable audible pop-up notifications
      --config string               config file (default is $HOME/.audiobooker.yaml)
      --debug                       debugging verbose output
      --dry-run                     Run parsing commands, without converting/binding, and display expected output
  -f, --file-pattern string         The output filename, can be a combination of literal values and patterns
  -j, --jobs int                    The number of concurrent transcoding process to run for conversion (don't exceed your cpu count) (default 1)
      --notify                      enable pop-up notifications
  -o, --output-directory string     The output directory for the final directory, can be combination of absolute values and path patterns
  -p, --path-pattern string         The pattern for metadata picked up via paths (starts from base of source-files-root)
      --scratch-files-path string   The location to generate the scratch directory
  -s, --source-files-root string    The path to directory of source files (must match path-pattern for metadata to work)
  -v, --verbose                     verbose output
      --verbose-transcode           Enable output of all ffmpeg commands/operations
```

### SEE ALSO

* [audiobooker batch](audiobooker_batch.md)	 - Perform batched operations on a pattern of directories for multiple audiobook binding

