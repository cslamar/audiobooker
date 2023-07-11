## audiobooker batch tag

Write tags to target audiobooks based on directory structures and path-pattern

### Synopsis

Write tags to target audiobooks based on directory structures and path-pattern.

```
audiobooker batch tag [flags]
```

### Options

```
  -h, --help   help for tag
```

### Options inherited from parent commands

```
      --config string               config file (default is $HOME/.audiobooker.yaml)
      --debug                       debugging verbose output
      --dry-run                     Run parsing commands, without converting/binding, and display expected output
  -f, --file-pattern string         The output filename, can be a combination of literal values and patterns
  -j, --jobs int                    The number of concurrent transcoding process to run for conversion (don't exceed your cpu count) (default 1)
  -o, --output-directory string     The output directory for the final directory, can be combination of absolute values and path patterns
  -p, --path-pattern string         The pattern for metadata picked up via paths (starts from base of source-files-root)
      --scratch-files-path string   The location to generate the scratch directory
  -s, --source-files-root string    The path to directory of source files (must match path-pattern for metadata to work)
  -v, --verbose                     verbose output
      --verbose-transcode           Enable output of all ffmpeg commands/operations
```

### SEE ALSO

* [audiobooker batch](audiobooker_batch.md)	 - Perform batched operations on a pattern of directories for multiple audiobook binding

