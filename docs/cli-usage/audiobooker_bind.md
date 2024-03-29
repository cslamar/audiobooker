## audiobooker bind

Combine multiple audio files into an M4B audiobook file

### Synopsis

The bind command, and its sub-commands, will perform actions to create a single audiobook.  This is usually used when you want to create a one off book, or aren't working through a full collection.

```
audiobooker bind [flags]
```

### Options

```
  -f, --file-pattern string         The output filename, can be a combination of literal values and patterns
  -h, --help                        help for bind
  -j, --jobs int                    The number of concurrent transcoding process to run for conversion (don't exceed your cpu count) (default 1)
  -o, --output-directory string     The output directory for the final directory, can be combination of absolute values and path patterns
  -p, --path-pattern string         The pattern for metadata picked up via paths
      --scratch-files-path string   The location to generate the scratch directory
  -s, --source-files-path string    The path to directory of source files (must match path-pattern for metadata to work)
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
* [audiobooker bind files](audiobooker_bind_files.md)	 - Bind audiobook using each file as a chapter
* [audiobooker bind from-tags](audiobooker_bind_from-tags.md)	 - Bind audiobook combining title tag of each file as chapter names
* [audiobooker bind split-chapters](audiobooker_bind_split-chapters.md)	 - Splits a single audio file into chapters using a fixed length
* [audiobooker bind tag](audiobooker_bind_tag.md)	 - Write tags to target audiobooks based on directory structures and path-pattern

