# Path Patterns

Tags are added either via the `--path-pattern` flag or `PATH_PATTERN` environment variable.  Supported path patterns are:

```text
Author      = "%a"
Genre       = "%g"
Narrator    = "%n"
ReleaseDate = "%y"
Series      = "%s"
SeriesPart  = "%p"
Title       = "%t"
```

Both `bind` and `batch` commands support tagging via path patterns.

When structuring path pattern arguments you must supply any hardcoded paths that aren't matched to a metadata parsing pattern.

#### Path Pattern Example

Using the following information as an example:

```text
      Author: Carl von Clausewitz
       Title: Volume 1
Release Year: 1903
 Series Name: On War
 Series Part: 1
```

With the filepath to the chapter files being:

```text
./media-src/files/1903/Carl von Clausewitz/On War/1/Volume 1
```

The matching path pattern, to pick up title, author, and release date would be:

```text
"./media-src/files/%y/%a/%s/%p/%t"
```

### Notes

* Though all commands use the same variables for the tags in the directories' paths, they employ the patterns in different ways.  Check the usage of each command for details.
