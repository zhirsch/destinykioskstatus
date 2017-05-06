#!/usr/bin/python

import json
import os
import sys
import tempfile
import urllib2
import zipfile

# Get the manifest urls.
req = urllib2.Request(
    "https://www.bungie.net//platform/Destiny/Manifest/",
    headers={'X-API-Key': sys.argv[1]},
)
resp = json.loads(urllib2.urlopen(req).read())
if resp['ErrorCode'] != 1:
    raise Exception("error: %s", resp)

with tempfile.TemporaryFile() as tf:
    # Download the zipped database.
    path = resp['Response']['mobileWorldContentPaths']['en']
    resp = urllib2.urlopen("https://www.bungie.net%s" %  path)
    while True:
        chunk = resp.read(16 << 10)
        if not chunk:
            break
        tf.write(chunk)

    # Unzip the database to the current directory.
    tf.seek(0)
    with zipfile.ZipFile(tf, 'r') as f:
        names = f.namelist()
        if len(names) != 1:
          raise Exception("too many entries: %s", names)
        f.extractall(path=os.path.dirname(sys.argv[2]))
        os.symlink(names[0], sys.argv[2])
