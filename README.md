cheerio
=========

A set of extensions to Python's pip.

Install
-------
Cheerio requires [Go](http://golang.org/doc/install).

`go install github.com/beyang/cheerio/...`

Usage
-----
Usage information: `cheerio -h`

### Examples
```
%> cheerio repo flask
http://github.com/mitsuhiko/flask
%> cheerio reqs flask-celery
pkg flask-celery uses (3):
  flask flask-script celery
and is used by (1):
    bundle-celery
```

### Regenerate data
The `cheerio reqs` subcommand uses a cached data file to get backward dependencies for PyPI packages.  This file is located in the `data/` directory.
It can be regenerated with `cheerio reqs-generate > <cache-file>`.  You can also specify the cache file optionally as in `cheerio reqs
-graphfile=<cache-file> <package-name>`.

Known issues
------------
* Does not correctly parse requirements for PyPI packages that contain multiple top-level packages (this is fairly rare)
* In querying PyPI metadata, cheerio fetches the last tarball/egg-file/zip-file listed by PyPI.  The ordering is alphanumeric, rather than by upload
  date.  Therefore, metadata for some packages (e.g., Flask) may be fetched from previous versions and thus be stale.

TODO
----
* mock pypi and add some unit tests
