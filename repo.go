package cheerio

import (
	"fmt"
	"regexp"
)

var homepageRegexp = regexp.MustCompile(`Home-page: (.+)\n`)
var repoPatterns = []*regexp.Regexp{
	regexp.MustCompile(`Home-page: (https?://github.com/(:?[^/\n\r]+)/(:?[^/\n\r]+))(:?/.*)?\s`),
	regexp.MustCompile(`Home-page: (https?://bitbucket.org/(:?[^/\n\r]+)/(:?[^/\n\r]+))(:?/.*)?\s`),
	regexp.MustCompile(`Home-page: (https?://code.google.com/p/(:?[^/\n\r]+))(:?/.*)?\s`),
}

var pkgInfoPattern = regexp.MustCompile(`(?:[^/]+/)*PKG\-INFO`)

func (p *PackageIndex) FetchSourceRepoURL(pkg string) (string, error) {
	b, err := p.FetchRawMetadata(pkg, pkgInfoPattern, pkgInfoPattern, pkgInfoPattern)
	if err != nil {
		// Try to fall back to hard-coded URLs
		if hardURL, in := pypiRepos[NormalizedPkgName(pkg)]; in {
			return hardURL, nil
		} else {
			return "", err
		}
	}
	rawMetadata := string(b)

	// Check PyPI
	for _, pattern := range repoPatterns {
		if match := pattern.FindStringSubmatch(rawMetadata); len(match) >= 1 {
			return match[1], nil
		}
	}

	// Try to fall back to hard-coded URLs
	if hardURL, in := pypiRepos[NormalizedPkgName(pkg)]; in {
		return hardURL, nil
	}

	// Return most informative error
	if match := homepageRegexp.FindStringSubmatch(rawMetadata); len(match) >= 1 {
		return "", fmt.Errorf("Could not parse repo URL from homepage: %s", match[1])
	}
	return "", fmt.Errorf("No homepage found in metadata: %s", rawMetadata)
}

var pypiRepos = map[string]string{
	"ajenti":                "git://github.com/Eugeny/ajenti",
	"ansible":               "git://github.com/ansible/ansible",
	"apache-libcloud":       "git://github.com/apache/libcloud",
	"autobahn":              "git://github.com/tavendo/AutobahnPython",
	"bottle":                "git://github.com/bottlepy/bottle",
	"celery":                "git://github.com/celery/celery",
	"chameleon":             "git://github.com/malthe/chameleon",
	"coverage":              "https://bitbucket.org/ned/coveragepy",
	"distribute":            "https://bitbucket.org/tarek/distribute",
	"django":                "git://github.com/django/django",
	"django-cms":            "git://github.com/divio/django-cms",
	"django-tastypie":       "git://github.com/toastdriven/django-tastypie",
	"djangocms-admin-style": "git://github.com/divio/djangocms-admin-style",
	"djangorestframework":   "git://github.com/tomchristie/django-rest-framework",
	"dropbox":               "git://github.com/sourcegraph/dropbox",
	"eve":                   "git://github.com/nicolaiarocci/eve",
	"fabric":                "git://github.com/fabric/fabric",
	"flask":                 "git://github.com/mitsuhiko/flask",
	"gevent":                "git://github.com/surfly/gevent",
	"gunicorn":              "git://github.com/benoitc/gunicorn",
	"httpie":                "git://github.com/jkbr/httpie",
	"httplib2":              "git://github.com/jcgregorio/httplib2",
	"itsdangerous":          "git://github.com/mitsuhiko/itsdangerous",
	"jinja2":                "git://github.com/mitsuhiko/jinja2",
	"kazoo":                 "git://github.com/python-zk/kazoo",
	"kombu":                 "git://github.com/celery/kombu",
	"lamson":                "git://github.com/zedshaw/lamson",
	"libcloud":              "git://github.com/apache/libcloud",
	"lxml":                  "git://github.com/lxml/lxml",
	"mako":                  "git://github.com/zzzeek/mako",
	"markupsafe":            "git://github.com/mitsuhiko/markupsafe",
	"matplotlib":            "git://github.com/matplotlib/matplotlib",
	"mimeparse":             "git://github.com/crosbymichael/mimeparse",
	"mock":                  "https://code.google.com/p/mock",
	"nltk":                  "git://github.com/nltk/nltk",
	"nose":                  "git://github.com/nose-devs/nose",
	"nova":                  "git://github.com/openstack/nova",
	"numpy":                 "git://github.com/numpy/numpy",
	"pandas":                "git://github.com/pydata/pandas",
	"pastedeploy":           "https://bitbucket.org/ianb/pastedeploy",
	"pattern":               "git://github.com/clips/pattern",
	"psycopg2":              "git://github.com/psycopg/psycopg2",
	"pyramid":               "git://github.com/Pylons/pyramid",
	"python-catcher":        "git://github.com/Eugeny/catcher",
	"python-dateutil":       "git://github.com/paxan/python-dateutil",
	"python-lust":           "git://github.com/zedshaw/python-lust",
	"pyyaml":                "git://github.com/yaml/pyyaml",
	"reconfigure":           "git://github.com/Eugeny/reconfigure",
	"repoze.lru":            "git://github.com/repoze/repoze.lru",
	"requests":              "git://github.com/kennethreitz/requests",
	"salt":                  "git://github.com/saltstack/salt",
	"scikit-learn":          "git://github.com/scikit-learn/scikit-learn",
	"scipy":                 "git://github.com/scipy/scipy",
	"sentry":                "git://github.com/getsentry/sentry",
	"setuptools":            "git://github.com/jaraco/setuptools",
	"sockjs-tornado":        "git://github.com/mrjoes/sockjs-tornado",
	"south":                 "https://bitbucket.org/andrewgodwin/south",
	"sqlalchemy":            "git://github.com/zzzeek/sqlalchemy",
	"ssh":                   "git://github.com/bitprophet/ssh",
	"tornado":               "git://github.com/facebook/tornado",
	"translationstring":     "git://github.com/Pylons/translationstring",
	"tulip":                 "git://github.com/sourcegraph/tulip",
	"twisted":               "git://github.com/twisted/twisted",
	"venusian":              "git://github.com/Pylons/venusian",
	"webob":                 "git://github.com/Pylons/webob",
	"webpy":                 "git://github.com/webpy/webpy",
	"werkzeug":              "git://github.com/mitsuhiko/werkzeug",
	"zope.interface":        "git://github.com/zopefoundation/zope.interface",
}
