import re
import sys
from httplib2 import Http
from sys import stderr
from os import path

def all_pkgs(uri):
    pkgs = []
    headers, body = Http().request(uri=('%s/simple' % uri), method='GET')
    matches = re.findall(r'\<a href=\'([A-Za-z0-9\._-]+)\'\>([A-Za-z0-9\._-]+)\</a\>\<br/\>', body)
    for match in matches:
        if len(match) != 2:
            stderr.write('Fatal: Unexpected number of matches')
            sys.exit()
        elif match[0] != match[1]:
            stderr.write('Error: pkg names do not match: %s != %s' % match)
        else:
            pkgs.append(match[0])
    return pkgs

def pkg_files(pypi_uri, pkg):
    files = []

    uri_path = '/simple/%s' % pkg
    uri = '%s%s' % (pypi_uri, uri_path)
    headers, body = Http().request(uri=uri, method='GET')
    matches = re.findall(r'\<a href="([/A-Za-z0-9\._-]+)#md5=[0-9a-z]+"[^\>]*\>([A-Za-z0-9\._-]+)\</a\>\<br/\>', body)
    if len(matches) == 0:
        # sys.stderr.write('Error: No matches found for %s\n' % body)
        return []

    for match in matches:
        files.append(path.normpath(path.join(uri_path, match[0])))
    return files



if __name__ == '__main__':
    pypi_uri = 'http://pypi.python.org'
    pkgs = all_pkgs(pypi_uri)
    for pkg in pkgs:
        files = pkg_files(pypi_uri, pkg)
        print files
    print len(pkgs)
