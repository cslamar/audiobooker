## audiobooker batch

Perform batched operations on a pattern of directories for multiple audiobook binding

### Synopsis

The batch command, and its sub-commands, will perform actions to create multiple audiobooks based on a collection of structured directories.  The way that the audiobooks are created is based on the pattern of the directory structure and sub-commands selected.  This is used when you want to convert a larger collection of books.

```
audiobooker batch [flags]
```

### Options

```
  -f, --file-pattern string         The output filename, can be a combination of literal values and patterns
  -h, --help                        help for batch
  -j, --jobs int                    The number of concurrent transcoding process to run for conversion (don't exceed your cpu count) (default 1)
  -o, --output-directory string     The output directory for the final directory, can be combination of absolute values and path patterns
  -p, --path-pattern string         The pattern for metadata picked up via paths (starts from base of source-files-root)
      --scratch-files-path string   The location to generate the scratch directory
  -s, --source-files-root string    The path to directory of source files (must match path-pattern for metadata to work)
      --verbose-transcode           Enable output of all ffmpeg commands/operations
```

### Options inherited from parent commands

```
      --alert           enable audible pop-up notifications
      --config string   config file (default is $HOME/.audiobooker.yaml)
      --debug           debugging verbose output
      --dry-run         Run parsing commands, without converting/binding, and display expected output
      --notify          enable pop-up notifications
  -v, --verbose         verbose output
```

### SEE ALSO

* [audiobooker](audiobooker.md)	 - Audiobook creation/manipulation application
* [audiobooker batch files](audiobooker_batch_files.md)	 - Bind audiobook using each file as a chapter
* [audiobooker batch from-tags](audiobooker_batch_from-tags.md)	 - Bind audiobook combining title tag of each file as chapter names
* [audiobooker batch split-chapters](audiobooker_batch_split-chapters.md)	 - Splits a single audio file into chapters using a fixed length
* [audiobooker batch tag](audiobooker_batch_tag.md)	 - Write tags to target audiobooks based on directory structures and path-pattern

