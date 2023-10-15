## audiobooker bind tag

Write tags to target audiobooks based on directory structures and path-pattern

### Synopsis

Write tags to target audiobooks based on directory structures and path-pattern.

```
audiobooker bind tag [flags]
```

### Options

```
  -h, --help   help for tag
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
  -p, --path-pattern string         The pattern for metadata picked up via paths
      --scratch-files-path string   The location to generate the scratch directory
  -s, --source-files-path string    The path to directory of source files (must match path-pattern for metadata to work)
  -v, --verbose                     verbose output
      --verbose-transcode           Enable output of all ffmpeg commands/operations
```

### SEE ALSO

* [audiobooker bind](audiobooker_bind.md)	 - Combine multiple audio files into an M4B audiobook file

