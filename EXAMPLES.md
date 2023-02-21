# Audioboker Command Examples

Below are a few examples, with summaries, on how to use Audiobooker

**Note:** `path-pattern` in `batch` works different from `bind`.  In `batch` the pattern starts **from** the path specified in `source-files-root`

### Split Into Fixed Length Chapters (with patterned output directory)

```shell
audiobooker bind split-chapters \
  --chapter-length 5 \
  --path-pattern "./media-src/%a/%s/%p/%t" \
  --output-directory "./ab/final/%a/%s/%p" \
  --source-files-path "./media-src/Carl von Clausewitz/On War/1/Volume 1"
```

### Create Audiobook From Structured Layout Compiling Chapters From Media Tags

```shell
audiobooker bind from-tags \
  --path-pattern "./media-src/%y/%a/%t" \
  --output-directory "./ab/final" \
  --source-files-path "./media-src/1903/Carl von Clausewitz/On War"
```

### Create Audiobook From Structured Layout creating one chapter per input file

```shell
audiobooker bind files \
  --path-pattern "./media-src/%y/%a/%t" \
  --output-directory "./ab/final" \
  --source-files-path "./media-src/1903/Carl von Clausewitz/On War"
```

### Create Audiobook From Structured Layout creating one chapter per input file with custom output path and filename

```shell
audiobooker bind files \
  --source-files-path "./media-src/files/chapter-by-file/1903/Carl von Clausewitz/On War" \
  --path-pattern "./media-src/files/chapter-by-file/%y/%a/%t" \
  --output-directory "./output/%a/%y/%t" \
  --file-pattern "%a - %t"
```

Would create the following directories and filename: `./output/Carl von Clausewitz/1903/On War/Carl von Clausewitz - On War.m4b`

### Batch a Collection of Books at Once, with One Chapter per File

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
